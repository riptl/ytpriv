package worker

import (
	"context"
	"time"
	log "github.com/sirupsen/logrus"
)

const vpsInterval = 3

// Expect store connected to Mongo and Redis
func Run(nThreads uint, ctxt context.Context) {
	var cancelFunc context.CancelFunc
	ctxt, cancelFunc = context.WithCancel(ctxt)
	wc := workerContext{
		ctxt: ctxt,
		errors: make(chan error),
		idleExists: make(chan bool),
		results: make(chan string),
	}

	activeRoutines := nThreads

	// Start routines
	for i := uint(0); i < nThreads; i++ {
		go wc.workRoutine()
	}

	// Count errors
	var errorTimes []time.Time

	// Idle queue timer
	var idleTimer <-chan time.Time

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
			log.WithField("vid", videoId).Debug("Visited video")
			videosLastInterval++

		// A routine exited
		case <-wc.idleExists:
			activeRoutines--
			idleTimer = time.After(3 * time.Second)

		// Print videos per second
		case <-vpsTimer:
			log.WithField("vps", videosLastInterval / vpsInterval).
				Info("Videos per second")
			videosLastInterval = 0
			// Respawn timer
			vpsTimer = time.After(vpsInterval * time.Second)

		// Time to respawn workers
		case <-idleTimer:
			delta := nThreads - activeRoutines
			log.Infof("%d threads idle because the queue is empty.", delta)
			log.Infof("Respawning %d threads", delta)
			for i := uint(0); i < delta; i++ {
				go wc.workRoutine()
			}

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
