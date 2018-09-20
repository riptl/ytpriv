package worker

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/store"
	"sync"
	"time"
)

// Writes newly found video IDs ("newIDs") and
// failed jobs ("failIDs") to the Redis queue.
// Results from "newIDs" are buffered and
// uploaded with "uploadNewVideoIDs"
func writeToQueue(
		newIDs <-chan []string,
		failIDs <-chan string,
		errors chan<- error,
		exit sync.WaitGroup) {
	exit.Add(1)

	var currentBatch []string
	var uploadTicker = time.NewTicker(1 * time.Second).C

	for { select {
		case id := <-failIDs:
			err := store.FailedVideoID(id)
			if err != nil {
				log.Errorf("Marking video \"%s\" as failed failed: %s", id, err.Error())
				errors <- err
			}

		case <-uploadTicker:
			if len(currentBatch) > 0 {
				uploadNewVideoIDs(currentBatch, errors)
				currentBatch = nil
			}

		case ids, more := <-newIDs:
			// Upload the last batch and exit
			if !more {
				uploadNewVideoIDs(currentBatch, errors)
				exit.Done()
				return
			}

			// Add the new IDs to the next batch
			currentBatch = append(currentBatch, ids...)
	}}
}

// Uploads a batch of newly found IDs without buffering
func uploadNewVideoIDs(ids []string, errors chan<- error) {
	now := time.Now()
	err := store.SubmitVideoIDs(ids)
	dur := time.Since(now)

	log.WithField("count", len(ids)).
		WithField("time", dur).
		Debug("Uploaded batch of found video IDs")

	if err != nil {
		log.Errorf("Pushing %d related video IDs failed: %s", len(ids), err.Error())
		errors <- err
	}
}

// Unpacks batches of new jobs from "batches to "jobs"
func unpackJobBatches(batches <-chan []string, jobs chan<- string) {
	for {
		batch, more := <-batches
		if !more { return }
		for _, id := range batch {
			jobs <- id
		}
	}
}

// Loads batches of new jobs to "jobsRaw" with size
// "batchSize". If the context is Done(), the "jobsRaw"
// channel is closed, causing the whole pipeline to stop.
func loadNewJobs(
		ctxt context.Context,
		batchSize uint,
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
		ids, err := store.GetScheduledVideoIDs(batchSize)
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
