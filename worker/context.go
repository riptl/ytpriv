package worker

import (
	"context"
	"github.com/terorie/yt-mango/data"
)

type workerContext struct{
	// Worker management
	ctxt context.Context

	// Sent when a worker gets
	// an error
	errors chan error

	// New videos (as batches from Redis)
	jobsRaw chan []string

	// New videos to process
	jobs chan string

	// Bulk buffer size
	bulkSize uint

	// Crawl results (buffered)
	results chan interface{}

	// Newly found IDs
	newIDs chan []string

	// Newly found IDs (as batches to Redis)
	newIDsRaw chan []string

	// Crawl result IDs
	resultIDs chan []string

	// Crawl fail IDs
	failIDs chan string

	// Crawl results in a batch
	// ready to be uploaded to DB
	resultBatches chan []data.Crawl

	// Sent whenever the
	// queue goes idle
	idle chan bool
}
