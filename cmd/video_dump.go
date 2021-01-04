package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

// TODO Wait for helper programs to exit

var videoDumpCmd = cobra.Command{
	Use:   "dump [video...]",
	Short: "Get a data set of videos",
	Long: "Dumps video metadata of specified video URLs.\n" +
		"If no videos are given as arguments, they are read from stdin.\n" +
		"The metadata is written to stdout as ndjson.",
	Args: cobra.ArbitraryArgs,
	Run:  cmdFunc(doVideoDump),
}

func init() {
	flags := videoDumpCmd.Flags()
	flags.UintP("num-related", "n", 0,
		"Number of videos to crawl from the related videos tab")
	flags.StringP("related-list", "r", "",
		"List of related videos that weren't crawled")
}

type videoDump struct {
	ctx            context.Context
	found          sync.Map
	queueLock      sync.Mutex
	queue          []string
	activeRoutines sync.WaitGroup
	count          int64
}

func doVideoDump(c *cobra.Command, args []string) (err error) {
	startTime := time.Now()

	videoIDs := make(chan string)
	videos := make(chan data.Video)

	flags := c.Flags()
	nRelated, err := flags.GetUint("num-related")
	if err != nil {
		return err
	}
	relatedList, err := flags.GetString("related-list")
	if err != nil {
		return err
	}

	var d videoDump
	go d.loadIDs(videoIDs, args, nRelated)

	// Spawn workers
	d.activeRoutines.Add(int(net.MaxWorkers))
	for i := 0; i < int(net.MaxWorkers); i++ {
		go d.videoDumpWorker(videos, videoIDs)
	}

	// Print stats
	go func() {
		for range time.Tick(time.Second) {
			logrus.WithField("count", atomic.LoadInt64(&d.count)).Info("Progress")
		}
	}()

	// ndjson output
	go func() {
		enc := json.NewEncoder(os.Stdout)
		for video := range videos {
			_ = enc.Encode(&video)
		}
	}()

	// Wait for routines to finish
	d.activeRoutines.Wait()
	close(videos)

	// Print success message
	logrus.Infof("Downloaded %d videos in %s",
		atomic.LoadInt64(&d.count),
		time.Since(startTime).String())

	// Write related videos list
	if relatedList != "" {
		f, err := os.Create(relatedList)
		if err != nil {
			logrus.WithError(err).Error("Failed to create related list")
			return err
		}
		defer f.Close()
		bf := bufio.NewWriter(f)
		defer bf.Flush()

		d.queueLock.Lock()
		for _, id := range d.queue {
			_, _ = bf.WriteString(id)
			err := bf.WriteByte('\n')
			if err != nil {
				logrus.WithError(err).Error("Failed to write related list")
				return err
			}
		}
		d.queueLock.Unlock()

		logrus.Info("Wrote related list")
	}

	return nil
}

func (d *videoDump) loadIDs(videoIDs chan<- string, args []string, nRelated uint) {
	var wg sync.WaitGroup
	defer close(videoIDs)
	defer wg.Wait()

	// Create argument channels
	jobs := make(chan string)
	go stdinOrArgs(jobs, args)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for job := range jobs {
			videoID, err := api.GetVideoID(job)
			if err != nil {
				logrus.Error(err)
				continue
			}
			videoIDs <- videoID
		}
	}()

	// Pop off queue
	for i := uint(0); i < nRelated; i++ {
		var s string
		var ok bool

		d.queueLock.Lock()
		if len(d.queue) > 0 {
			ok = true
			s = d.queue[0]
			d.queue = d.queue[1:]
		} else {
			ok = false
		}
		d.queueLock.Unlock()

		if !ok {
			logrus.Warnf("Not enough videos found (%d missing)",
				nRelated-i-1)
			return
		}

		videoIDs <- s
	}
}

func (d *videoDump) videoDumpWorker(videos chan<- data.Video, videoIDs <-chan string) {
	defer d.activeRoutines.Done()
	for videoID := range videoIDs {
		v, ok := d.videoDumpSingle(videoID)
		if ok {
			videos <- v
			atomic.AddInt64(&d.count, 1)
		}
	}
}

func (d *videoDump) videoDumpSingle(videoId string) (v data.Video, ok bool) {
	// Download video info
	req := api.GrabVideo(videoId)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := net.Client.Do(req, res)
	if err != nil {
		logrus.WithError(err).
			WithField("id", videoId).
			Error("Failed to download video")
		return
	}

	v.ID = videoId

	// Parse video
	err = api.ParseVideo(&v, res)
	if err != nil {
		logrus.WithError(err).
			WithField("id", videoId).
			Error("Failed to parse video")
		return
	}

	var newIDs []string
	for _, vid := range v.RelatedVideos {
		if _, loaded := d.found.LoadOrStore(vid.ID, true); !loaded {
			newIDs = append(newIDs, vid.ID)
		}
	}

	d.queueLock.Lock()
	d.queue = append(d.queue, newIDs...)
	d.queueLock.Unlock()

	// Send video to writer
	logrus.WithField("id", videoId).Debug("Got video")

	return v, true
}
