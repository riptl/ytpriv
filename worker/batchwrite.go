package worker

import (
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/store"
	"time"
)

// Uploads batches to Mongo

func (c *workerContext) batchUploader() {
	var timeout <-chan time.Time
	for { select {
		case <-c.ctxt.Done():
			timeout = time.After(1 * time.Second)
			return

		case batch := <-c.resultBatches:
			start := time.Now()
			err := store.SubmitCrawls(batch)
			dur := time.Since(start)

			// Upload to Mongo
			log.WithField("count", len(batch)).
				WithField("time", dur).
				Info("Uploaded batch of videos")

			if err != nil {
				log.Errorf("Uploading crawl of %d videos failed: %s", len(batch), err.Error())
				c.errors <- err
			}

		case <-timeout:
			return
	}}
}
