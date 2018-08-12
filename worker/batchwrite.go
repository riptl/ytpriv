package worker

import (
	"github.com/terorie/yt-mango/data"
	"time"
	"github.com/terorie/yt-mango/store"
	log "github.com/sirupsen/logrus"
	"sync"
)

// Uploads batches to Mongo

func uploadResults(
		resultBatches <-chan []data.Crawl,
		errors chan<- error,
		exit sync.WaitGroup) {
	exit.Add(1)

	for {
		batch, more := <-resultBatches
		if !more {
			exit.Done()
			return
		}

		start := time.Now()
		err := store.SubmitCrawls(batch)
		dur := time.Since(start)

		// Upload to Mongo
		log.WithField("count", len(batch)).
		WithField("time", dur).
		Debug("Uploaded batch of videos")

		if err != nil {
			log.Errorf("Uploading crawl of %d videos failed: %s", len(batch), err.Error())
			errors <- err
		}
	}
}
