package worker

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/store"
)

func (c *workerContext) workRoutine() {
	for {
		// Check if routine should exit
		select {
			case <-c.ctxt.Done(): break
			default:
		}

		// TODO Move video back to wait queue if processing failed

		videoId := <-c.jobs
		req := apis.Main.GrabVideo(videoId)
		res, err := net.Client.Do(req)
		if err != nil {
			log.Errorf("Failed to download video \"%s\": %s", videoId, err.Error())
			c.errors <- err
		}

		var v data.Video
		v.ID = videoId
		var result interface{}

		next, err := apis.Main.ParseVideo(&v, res)
		if err == api.VideoUnavailable {
			log.Debugf("Video is unavailable: %s", videoId)
			result = data.CrawlError{ uint(api.VideoUnavailable), time.Now() }
		} else if err != nil {
			log.Errorf("Parsing video \"%s\" failed: %s", videoId, err.Error())
			c.errors <- err
		} else {
			result = data.Crawl{ &v, time.Now() }
		}

		err = store.SubmitCrawl(result)
		if err != nil {
			log.Errorf("Uploading crawl of video \"%s\" failed: %s", videoId, err.Error())
			c.errors <- err
		}

		if len(next) > 0 {
			err = store.SubmitVideoIDs(next)
			if err != nil {
				log.Errorf("Pushing related video IDs of video \"%s\" failed: %s", videoId, err.Error())
				c.errors <- err
			}
		}

		c.results <- videoId
	}
}
