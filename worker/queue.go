package worker

import (
	"github.com/terorie/yt-mango/store"
	log "github.com/sirupsen/logrus"
	"time"
)

// Queue handler:
// Reads and writes to queue in the background

// TODO Handle errors
func (c *workerContext) handleQueueWrite() {
	var timeOut <-chan time.Time
	for { select {
		case <-c.ctxt.Done():
			timeOut = time.After(1 * time.Second)

		case ids := <-c.resultIDs:
			c.queueResultIDs(ids)

		case id := <-c.failIDs:
			c.queueFailID(id)

		case ids := <-c.newIDsRaw:
			c.queueNewIDs(ids)

		case <-timeOut:
			return
	}}
}

// Packs batches of new video IDs from
// "newIDs" to "newIDsRaw".
// Sends the new IDs to Redis every second.
func (c *workerContext) handleQueueWriteHelper() {
	pendingShutdown := false

	timeOut := time.After(1 * time.Second)
	var idBuf []string

	for { select {
		case <-c.ctxt.Done():
			pendingShutdown = true

		case ids := <-c.newIDs:
			idBuf = append(idBuf, ids...)

		case <-timeOut:
			c.newIDsRaw <- idBuf
			idBuf = nil
			if pendingShutdown { return }
			timeOut = time.After(1 * time.Second)
	}}
}

func (c *workerContext) queueResultIDs(ids []string) {
	err := store.DoneVideoIDs(ids)
	if err != nil {
		log.Errorf("Marking %d videos as done failed: %s", len(ids), err.Error())
		c.errors <- err
	}
}

func (c *workerContext) queueFailID(id string) {
	err := store.FailedVideoID(id)
	if err != nil {
		log.Errorf("Marking video \"%s\" as failed failed: %s", id, err.Error())
		c.errors <- err
	}
}

func (c *workerContext) queueNewIDs(ids []string) {
	err := store.SubmitVideoIDs(ids)
	if err != nil {
		log.Errorf("Pushing %d related video IDs failed: %s", len(ids), err.Error())
		c.errors <- err
	}
}

// Unpacks batches of new jobs from
// "jobsRaw" to "jobs"
func (c *workerContext) handleQueueReceiveHelper() {
	for { select {
		case <-c.ctxt.Done():
			return

		case batch := <-c.jobsRaw:
			for _, id := range batch {
				c.jobs <- id
			}
	}}
}

func (c *workerContext) handleQueueReceive() {
	for { select {
		case <-c.ctxt.Done():
			return

		default:
			ids, err := store.GetScheduledVideoIDs(c.bulkSize)
			if err != nil && err.Error() != "redis: nil" {
				log.Error("Queue error: ", err.Error())
				c.errors <- err
			}
			if len(ids) == 0 {
				// Queue is empty
				log.Info("No jobs on queue. Waiting 1 second.")
				time.Sleep(1 * time.Second)
				continue
			}
			c.jobsRaw <- ids
	}}
}
