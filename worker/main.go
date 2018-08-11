package worker

import (
	"os"
	"context"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/viperstruct"
)

const vpsInterval = 3

func Run(rootCtxt context.Context, firstId string) {
	// Read config
	viper.SetDefault("myname", "")
	viper.SetDefault("connections", 32)
	viper.SetDefault("batchsize", 16)
	viper.SetDefault("batches", 4)

	var conf struct{
		Connections uint `viper:"connections,optional"`
		BulkWriteSize uint `viper:"batchsize,optional"`
		Batches uint `viper:"batches,optional"`
		MyName string `viper:"myname,optional"`
	}
	err := viperstruct.ReadConfig(&conf)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
	if len(conf.MyName) > 10 {
		log.Errorf("Name \"%s\" is longer than 10 bytes!")
		os.Exit(1)
	}

	var cancelFunc context.CancelFunc
	ctxt, cancelFunc := context.WithCancel(rootCtxt)

	// Channels
	chanSize := 2 * conf.Connections
	errors := make(chan error)
	bulkSize := conf.BulkWriteSize
	results := make(chan interface{}, chanSize)
	resultBatches := make(chan []data.Crawl, conf.Batches)

	// Receive jobs
	jobsRaw := make(chan []string, 2)
	jobs := make(chan string, chanSize)
	go handleQueueReceive(ctxt, bulkSize, jobsRaw, errors)
	go handleQueueReceiveHelper(jobsRaw, jobs)

	// Write results
	newIDs := make(chan []string, chanSize)
	newIDsRaw := make(chan []string, 2)
	failIDs := make(chan string, chanSize)
	go handleQueueWrite(newIDsRaw, failIDs, errors)
	go handleQueueWriteHelper(newIDs, newIDsRaw)

	// Result handler
	go handleResults(conf.BulkWriteSize, results, resultBatches, failIDs)
	// Data uploader
	go batchUploader(resultBatches, errors)

	if firstId != "" {
		newIDsRaw <- []string{firstId}
		log.Infof("Pushed first ID \"%s\".", firstId)
	}

	// Start workers
	activeRoutines := conf.Connections
	onExit := make(chan struct{})
	for i := uint(0); i < conf.Connections; i++ {
		go workRoutine(conf.MyName, jobs, results, newIDs, errors, onExit)
	}

	// Count errors
	var errorTimes []time.Time

	// Collect info from routines
	for { select {
		case <-ctxt.Done():
			log.Info("Requested cancellation.")
			return

		// Worker routine exited
		case <-onExit:
			activeRoutines--
			if activeRoutines == 0 {
				// No results available anymore
				close(results)
			}

		// Rescan and drop old errors
		case <-errors:
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
			if len(errorTimes) > int(conf.Connections / 2) {
				log.Error("Exiting because of too many errors.")
				cancelFunc()
				return
			}
	}}
}
