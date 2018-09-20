package worker

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/api"
)

func workRoutine(
		crawlerName string,
		jobs <-chan string,
		results chan<- interface{},
		newIDs chan<- []string,
		errors chan<- error,
		notifyShutdown chan<- struct{}) {
	for {
		// TODO Move video back to wait queue if processing failed
		videoId, more := <-jobs

		// No jobs, channel closed
		if !more {
			notifyShutdown <- struct{}{}
			return
		}

		req := apis.Main.GrabVideo(videoId)
		res, err := net.Client.Do(req)
		if err != nil {
			log.Errorf("Failed to download video \"%s\": %s", videoId, err.Error())
			errors <- err
		}

		var v data.Video
		v.ID = videoId
		var result interface{}

		err = apis.Main.ParseVideo(&v, res)
		if err == api.VideoUnavailable {
			log.Debugf("Video is unavailable: %s", videoId)
			result = data.CrawlError{
				VideoId: videoId,
				Err: api.VideoUnavailable,
				VisitedTime: time.Now(),
			}
		} else if err != nil {
			log.Errorf("Parsing video \"%s\" failed: %s", videoId, err.Error())
			errors <- err
		} else {
			result = data.Crawl{
				Video: &v,
				VisitedTime: time.Now(),
				CrawlerName: crawlerName,
			}
		}

		results <- result

		if len(v.Related) > 0 {
			newIDs <- v.Related
		}
	}
}
