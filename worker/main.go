package worker

import (
	"context"
	"time"
	log "github.com/sirupsen/logrus"
)

// Expect store connected to Mongo and Redis
func Run(nThreads uint, ctxt context.Context) {
	var cancelFunc context.CancelFunc
	ctxt, cancelFunc = context.WithCancel(ctxt)
	wc := workerContext{
		ctxt: ctxt,
		errors: make(chan error),
		idleExists: make(chan bool),
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

	// Maintain routines
	for {
		select {
		case <-ctxt.Done():
			log.Info("Requested cancellation.")
			return

		case <-wc.idleExists:
			// A routine exited
			activeRoutines--
			idleTimer = time.After(3 * time.Second)

		case <-idleTimer:
			// Time to respawn workers
			delta := nThreads - activeRoutines
			log.Infof("%d threads idle because the queue is empty.", delta)
			log.Infof("Respawning %d threads", delta)
			for i := uint(0); i < delta; i++ {
				go wc.workRoutine()
			}

		case <-wc.errors:
			// Rescan and drop old errors
			var newErrorTimes []time.Time
			for _, t := range errorTimes {
				if t.Sub(time.Now()) < 10 * time.Second {
					newErrorTimes = append(newErrorTimes, t)
				}
			}
			errorTimes = newErrorTimes

			// New error received
			errorTimes = append(errorTimes, time.Now())

			// If too many errors exit
			if len(errorTimes) > 10 {
				log.Error("Exiting because of too many errors.")
				cancelFunc()
				return
			}
		}
	}
}
