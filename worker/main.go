package worker

import (
	"context"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/store"
)

const vpsInterval = 3

// Expect store connected to Mongo and Redis
func Run(nThreads uint, ctxt context.Context) {
	var cancelFunc context.CancelFunc
	ctxt, cancelFunc = context.WithCancel(ctxt)
	wc := workerContext{
		ctxt: ctxt,
		errors: make(chan error),
		jobs: make(chan string),
		idle: make(chan bool),
		results: make(chan string),
	}

	// Start routines
	for i := uint(0); i < nThreads; i++ {
		go wc.workRoutine()
	}

	// Start loading queue
	go wc.fetchJobs()

	// Count errors
	var errorTimes []time.Time

	// Videos per second timer
	var videosLastInterval = 0
	var vpsTimer = time.After(vpsInterval * time.Second)

	// Maintain routines
	for {
		select {
		case <-ctxt.Done():
			log.Info("Requested cancellation.")
			return

		// Video crawled
		case videoId := <-wc.results:
			err := store.DoneVideoID(videoId)
			if err != nil {
				log.Errorf("Marking video \"%s\" as done failed: %s", videoId, err.Error())
				cancelFunc()
				return
			}
			log.WithField("vid", videoId).Debug("Visited video")
			videosLastInterval++

		// A routine exited
		case <-wc.idle:
			log.Info("No jobs on queue. Waiting 10 seconds.")
			time.Sleep(10 * time.Second)
			go wc.fetchJobs()

		// Print videos per second
		case <-vpsTimer:
			log.WithField("vps", videosLastInterval / vpsInterval).
				Info("Videos per second")
			videosLastInterval = 0
			// Respawn timer
			vpsTimer = time.After(vpsInterval * time.Second)

		// Rescan and drop old errors
		case <-wc.errors:
			var newErrorTimes []time.Time
			for _, t := range errorTimes {
				if t.Sub(time.Now()) < 5 * time.Second {
					newErrorTimes = append(newErrorTimes, t)
				}
			}
			errorTimes = newErrorTimes

			// New error received
			errorTimes = append(errorTimes, time.Now())

			// If too many errors exit
			if len(errorTimes) > int(nThreads / 2) {
				log.Error("Exiting because of too many errors.")
				cancelFunc()
				return
			}
		}
	}
}

func (c *workerContext) fetchJobs() {
	for {
		select {
		case <-c.ctxt.Done():
			return
		default:
			videoId, err := store.GetScheduledVideoID()
			if err != nil && err.Error() != "redis: nil" {
				log.Error("Queue error: ", err.Error())
				c.errors <- err
			}
			if videoId == "" {
				// Queue is empty, break
				c.idle <- true
				return
			}

			c.jobs <- videoId
		}
	}
}
