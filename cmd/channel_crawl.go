package cmd

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
	"sync"
	"sync/atomic"
	"time"
)

var channelCrawlCmd = cobra.Command{
	Use: "crawl <channel ID>",
	Short: "Crawl channel IDs",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doChannelCrawl),
}

type channelDump struct{
	found          *bigcache.BigCache
	queueLock      sync.Mutex
	queue          []string
	activeRoutines sync.WaitGroup
	count          int64
	requestCount   int64
}

func doChannelCrawl(_ *cobra.Command, args []string) (err error) {
	startIDs := args

	for i, url := range startIDs {
		startIDs[i], err = api.GetVideoID(url)
		if err != nil {
			return err
		}
	}

	bigCache, err := bigcache.NewBigCache(bigcache.Config{
		Shards:             1024,
		CleanWindow:        0,
		MaxEntriesInWindow: 1000 * 10 * 60,
		MaxEntrySize:       500,
		HardMaxCacheSize:   1e8, // 100 MB
	})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to init bigcache")
	}

	d := channelDump{
		found: bigCache,
		queue: startIDs,
	}

	// Spawn workers
	d.activeRoutines.Add(int(net.MaxWorkers))
	for i := 0; i < int(net.MaxWorkers); i++ {
		go d.videoDumpWorker()
	}

	ctx, cancel := context.WithCancel(context.Background())
	go d.log(ctx)

	// Wait for routines to finish
	d.activeRoutines.Wait()

	cancel()

	return nil
}

func (d *channelDump) log(c context.Context) {
	lastN := int64(0)
	lastReq := int64(0)
	ticks := time.Tick(10 * time.Second)
	for {
		select {
		case <-c.Done():
			return
		case <-ticks:
			nowN := atomic.LoadInt64(&d.count)
			rate := float64(nowN - lastN) / 10
			lastN = nowN

			nowReq := atomic.LoadInt64(&d.requestCount)
			rateReq := float64(nowReq - lastReq) / 10
			lastReq = nowReq

			logrus.WithFields(logrus.Fields{
				"rate": rate,
				"req_rate": rateReq,
				"total": nowN,
			}).Info("Stats")
		}
	}
}

func (d *channelDump) getNextID() (s string, ok bool) {
	d.queueLock.Lock()
	defer d.queueLock.Unlock()

	if len(d.queue) > 1000000 {
		// Queue too long do a snibbening
		logrus.Warning("Truncating big queue")
		d.queue = d.queue[500000:]
	}

	if len(d.queue) > 0 {
		ok = true
		s = d.queue[0]
		d.queue = d.queue[1:]
	} else {
		ok = false
	}
	return
}

func (d *channelDump) videoDumpWorker() {
	defer d.activeRoutines.Done()

	for {
		videoId, ok := d.getNextID()
		if !ok {
			return
		}

		d.videoDumpSingle(videoId)
	}
}

func (d *channelDump) videoDumpSingle(videoId string) {
	// Download video info
	req := apis.Main.GrabVideo(videoId)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err := net.Client.Do(req, res)
	atomic.AddInt64(&d.requestCount, 1)
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
		if vid.UploaderID == "" {
			logrus.WithField("vid", vid.ID).
				Error("Got empty uploader ID")
			continue
		}

		// Throw uploader in cache
		_, getErr := d.found.Get(vid.UploaderID)
		if getErr != bigcache.ErrEntryNotFound {
			// Uploader already known
			continue
		}
		_ = d.found.Set(vid.UploaderID, nil)

		atomic.AddInt64(&d.count, 1)
		fmt.Printf("https://www.youtube.com/channel/%s\n", vid.UploaderID)

		// Prepare to crawl video
		newIDs = append(newIDs, vid.ID)
	}

	d.queueLock.Lock()
	defer d.queueLock.Unlock()
	d.queue = append(d.queue, newIDs...)

	return
}
