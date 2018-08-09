package worker

import (
	"github.com/terorie/yt-mango/store"
	log "github.com/sirupsen/logrus"
	"time"
)

// Queue handler:
// Reads and writes to queue in the background

func (c *workerContext) handleQueue() {
	for { select {
		case <-c.ctxt.Done():
			c.handleQueueNoFetch(1 * time.Second)
			return

		case id := <-c.resultIDs:
			store.DoneVideoID(id)

		case id := <-c.failIDs:
			store.FailedVideoID(id)

		default:
			videoId, err := store.GetScheduledVideoID()
			if err != nil && err.Error() != "redis: nil" {
				log.Error("Queue error: ", err.Error())
				c.errors <- err
			}
			if videoId == "" {
				// Queue is empty
				log.Info("No jobs on queue. Waiting 1 second.")
				c.handleQueueNoFetch(1 * time.Second)
			}

			c.jobs <- videoId
	}}
}

func (c *workerContext) handleQueueNoFetch(duration time.Duration) {
	timeOut := time.After(duration)
	for { select {
		case id := <-c.resultIDs:
			store.DoneVideoID(id)

		case id := <-c.failIDs:
			store.FailedVideoID(id)

		case <-timeOut:
			return
	}}
}
