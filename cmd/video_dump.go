package cmd

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var videoDumpCmd = cobra.Command{
	Use: "dump <n> <video>",
	Short: "Get a data set of videos",
	Long: "Loads n video metadata files to the\n" +
		"specified path starting at the root video.\n" +
		"The videos are written to stdout as ndjson.",
	Args: cobra.MinimumNArgs(2),
	Run: cmdFunc(doVideoDump),
}

type videoDump struct{
	ctx            context.Context
	out            chan data.Video
	left           int32
	found          sync.Map
	queueLock      sync.Mutex
	queue          []string
	foundAll       int32
	activeRoutines sync.WaitGroup
}

func doVideoDump(_ *cobra.Command, args []string) (err error) {
	startTime := time.Now()

	nStr := args[0]
	startIDs := args[1:]

	toDownload, err := strconv.ParseUint(nStr, 10, 31)
	if err != nil { return }

	for i, url := range startIDs {
		startIDs[i], err = api.GetVideoID(url)
		if err != nil { return }
	}

	d := videoDump{
		out:      make(chan data.Video),
		left:     int32(toDownload),
		queue:    startIDs,
		foundAll: 0,
	}
	go d.writer()

	// Spawn workers
	for i := 0; i < int(net.MaxWorkers); i++ {
		d.activeRoutines.Add(1)
		go d.videoDumpWorker()
	}

	// Print stats
	go func() {
		for range time.NewTicker(time.Second).C {
			_left := atomic.LoadInt32(&d.left)
			if _left != 0 {
				logrus.WithField("left", _left).
					Infof("%d videos left", _left)
			} else {
				break
			}
		}
	}()

	// Wait for routines to finish
	d.activeRoutines.Wait()
	close(d.out)

	// Print success message
	logrus.Infof("Downloaded %d videos in %s",
		toDownload,
		time.Since(startTime).String())

	return nil
}

func (d *videoDump) writer() {
	enc := json.NewEncoder(os.Stdout)
	for vid := range d.out {
		err := enc.Encode(&vid)
		if err != nil {
			logrus.WithError(err).Fatal("Output stream aborted")
		}
	}
}

func (d *videoDump) getNextID() (s string, ok bool) {
	d.queueLock.Lock()
	defer d.queueLock.Unlock()

	if len(d.queue) > 0 {
		ok = true
		s = d.queue[0]
		d.queue = d.queue[1:]
	} else {
		ok = false
	}
	return
}

func (d *videoDump) videoDumpWorker() {
	defer d.activeRoutines.Done()

	for {
		// Check if all videos loaded
		if atomic.LoadInt32(&d.left) == 0 { return }

		videoId, ok := d.getNextID()
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Update videos left counter
		if atomic.AddInt32(&d.left, -1) <= 0 { return }

		if !d.videoDumpSingle(videoId) {
			// Increment videos left counter back again
			atomic.AddInt32(&d.left, +1)
		}
	}
}

func (d *videoDump) videoDumpSingle(videoId string) (success bool) {
	success = false

	// Download video info
	req := apis.Main.GrabVideo(videoId)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := net.Client.Do(req, res)
	if err != nil {
		logrus.WithError(err).
			WithField("id", videoId).
			Error("Failed to download video")
		return
	}

	var v data.Video
	v.ID = videoId

	// Parse video
	err = apis.Main.ParseVideo(&v, res)
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

	// Add newly found video IDs
	if atomic.LoadInt32(&d.foundAll) == 0 {
		d.queueLock.Lock()
		d.queue = append(d.queue, newIDs...)
		d.queueLock.Unlock()
	}

	// Send video to writer
	d.out <- v
	logrus.WithField("id", videoId).Debug("Got video")

	success = true
	return
}
