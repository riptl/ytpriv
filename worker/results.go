package worker

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/data"
)

// Collect results from wc.results,
// write batches to Redis and Mongo
func (c *workerContext) handleResults() {
	resultBuf := make([]data.Crawl, c.bulkSize)
	resultBufPtr := uint(0)
	exitNow := false

	// Videos per second timer
	var videosLastInterval = 0
	var vpsTimer = time.After(vpsInterval * time.Second)

	for { select{
		case <-c.ctxt.Done():
			exitNow = true

		case result := <-c.results:
			switch result.(type) {
			case data.Crawl:
				vid := result.(data.Crawl)

				// Log info
				log.WithField("vid", vid.Video.ID).Debug("Visited video")
				videosLastInterval++

				resultBuf[resultBufPtr] = vid
				resultBufPtr++
				// Buffer full
				if resultBufPtr == c.bulkSize || exitNow {
					c.resultBatches <- resultBuf
					resultBufPtr = 0
				}
			case data.CrawlError:
				ce := result.(data.CrawlError)

				// Log error
				log.Errorf("Marking video \"%s\" as done failed: %s",
					ce.VideoId, ce.Err.Error())

				c.failIDs <- ce.VideoId
			}

		// Print videos per second
		case <-vpsTimer:
			log.WithField("vps", videosLastInterval / vpsInterval).
				Info("Videos per second")
			videosLastInterval = 0
			// Respawn timer
			vpsTimer = time.After(vpsInterval * time.Second)
	}}
}
