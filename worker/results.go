package worker

import (
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/data"
)

// Collect results from wc.results,
// write batches to Mongo
func collectResults(
		bulkSize uint,
		results <-chan interface{},
		resultBatches chan<- []data.Crawl,
		failIDs chan<- string) {
	resultBuf := make([]data.Crawl, bulkSize)
	resultBufPtr := uint(0)

	// Videos per second timer
	totalVideoCount := uint(0)
	videosLastInterval := uint(0)
	vpsTicker := time.NewTicker(vpsInterval * time.Second).C

	for { select{
		case result, more := <-results:
			// No more new results
			if !more {
				resultBatches <- resultBuf
				close(resultBatches)
				return
			}

			switch result.(type) {
			case data.Crawl:
				vid := result.(data.Crawl)

				// Log info
				videosLastInterval++

				resultBuf[resultBufPtr] = vid
				resultBufPtr++
				// Buffer full
				if resultBufPtr == bulkSize {
					resultBatches <- resultBuf
					resultBufPtr = 0
				}
			case data.CrawlError:
				ce := result.(data.CrawlError)

				// Log error
				log.Errorf("Marking video \"%s\" as done failed: %s",
					ce.VideoId, ce.Err.Error())

				failIDs <- ce.VideoId
			}

		// Print videos per second
		case <-vpsTicker:
			totalVideoCount += videosLastInterval
			log.WithField("perSecond", videosLastInterval / vpsInterval).
				WithField("total", totalVideoCount).
				Info("Got new videos")
			videosLastInterval = 0
	}}
}
