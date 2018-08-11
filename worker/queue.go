package worker

import (
	"github.com/terorie/yt-mango/store"
	log "github.com/sirupsen/logrus"
	"time"
	"context"
)

// Queue handler:
// Reads and writes to queue in the background

// TODO Handle errors
func handleQueueWrite(
		newIDsRaw <-chan []string,
		failIDs <-chan string,
		errors chan<- error) {
	for { select {
		case id := <-failIDs:
			err := store.FailedVideoID(id)
			if err != nil {
				log.Errorf("Marking video \"%s\" as failed failed: %s", id, err.Error())
				errors <- err
			}

		case ids, more := <-newIDsRaw:
			if !more { return }
			err := store.SubmitVideoIDs(ids)
			if err != nil {
				log.Errorf("Pushing %d related video IDs failed: %s", len(ids), err.Error())
				errors <- err
			}
	}}
}

// Packs batches of new video IDs from
// "newIDs" to "newIDsRaw".
// Sends the new IDs to Redis every second.
func handleQueueWriteHelper(newIDs <-chan []string, newIDsRaw chan<- []string) {
	var idBuf []string
	var lastPersist time.Time

	for {
		ids, more := <-newIDs

		// Upload the last batch and exit
		if !more {
			newIDsRaw <- idBuf
			return
		}

		// Last upload too long ago
		// Upload next IDs
		if time.Since(lastPersist) > 1 * time.Second {
			newIDsRaw <- idBuf
			idBuf = nil
			lastPersist = time.Now()
		}

		idBuf = append(idBuf, ids...)
	}
}

// Unpacks batches of new jobs from
// "jobsRaw" to "jobs"
func handleQueueReceiveHelper(jobsRaw <-chan []string, jobs chan<- string) {
	for {
		batch, more := <-jobsRaw
		if !more { return }
		for _, id := range batch {
			jobs <- id
		}
	}
}

func handleQueueReceive(
		ctxt context.Context,
		bulkSize uint,
		jobsRaw chan<- []string,
		errors chan<- error) {
	for {
		select {
		case <-ctxt.Done():
			// Stop signal received, no more jobs
			close(jobsRaw)
			return
		default:
		}

		// Request videos from Redis with time
		start := time.Now()
		ids, err := store.GetScheduledVideoIDs(bulkSize)
		dur := time.Since(start)

		// Redis error or no results
		if err != nil && err.Error() != "redis: nil" {
			log.Error("Queue error: ", err.Error())
			errors <- err
		}

		// Queue is empty
		if len(ids) == 0 {
			log.Info("No jobs on queue. Waiting 1 second.")
			time.Sleep(1 * time.Second)
			continue
		}

		// Log receive
		log.WithField("count", len(ids)).
			WithField("time", dur).
			Debug("Received batch of jobs")

		jobsRaw <- ids
	}
}
