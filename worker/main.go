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

func Run(ctxt context.Context, firstId string) {
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

	var c workerContext
	c.foundBy = conf.MyName

	var cancelFunc context.CancelFunc
	c.ctxt, cancelFunc = context.WithCancel(ctxt)

	// Channels
	chanSize := 2 * conf.Connections
	c.errors = make(chan error)
	c.jobsRaw = make(chan []string, 2)
	c.jobs = make(chan string, chanSize)
	c.bulkSize = conf.BulkWriteSize
	c.results = make(chan interface{}, chanSize)
	c.newIDs = make(chan []string, chanSize)
	c.newIDsRaw = make(chan []string, 2)
	c.failIDs = make(chan string, chanSize)
	c.resultBatches = make(chan []data.Crawl, conf.Batches)
	c.idle = make(chan bool)

	// Redis handler
	go c.handleQueueReceive()
	go c.handleQueueReceiveHelper()
	go c.handleQueueWrite()
	go c.handleQueueWriteHelper()
	// Result handler
	go c.handleResults()
	// Data uploader
	go c.batchUploader()

	if firstId != "" {
		c.newIDsRaw <- []string{firstId}
		log.Infof("Pushed first ID \"%s\".", firstId)
	}

	// Start workers
	for i := uint(0); i < conf.Connections; i++ {
		go c.workRoutine()
	}

	// Count errors
	var errorTimes []time.Time

	// Collect info from routines
	for { select {
		case <-ctxt.Done():
			log.Info("Requested cancellation.")
			return

		// Rescan and drop old errors
		case <-c.errors:
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
