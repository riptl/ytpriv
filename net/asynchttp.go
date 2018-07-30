package net

import (
	"net/http"
	"sync/atomic"
	"time"
)

// Max number of HTTP workers
var MaxWorkers uint32 = 4
// Current number of HTTP workers
// atomic variable, don't use directly
var activeWorkers int32

// Kill a worker routine if it
// doesn't get any jobs after "timeOut"
const timeOut = 10 * time.Second

// Result of the HTTP request
type JobResult struct {
	// HTTP Response (can be nil)
	Res *http.Response
	// HTTP error (can be nil)
	Err error
	// data parameter from DoAsyncHTTP
	ReqData interface{} // job.data
}

type job struct {
	req *http.Request
	c chan JobResult
	data interface{}
}

// Job queue
var jobs = make(chan job)

// Enqueue a new HTTP request and send the result to "c" (send to "c" guaranteed)
// Additional data like an ID can be passed in "data" to be returned with "c"
func DoAsyncHTTP(r *http.Request, c chan JobResult, data interface{}) {
	newJob := job{r, c, data}
	select {
		// Try to send to the channel and
		// see if an idle worker picks the job up
		case jobs <- newJob:
			break

		// Every routine is busy
		default:
			if atomic.LoadInt32(&activeWorkers) < int32(MaxWorkers) {
				// Another thread is allowed to spawn
				// TODO Race condition here: DoAsyncHTTP is not thread safe!
				atomic.AddInt32(&activeWorkers, 1)
				go asyncHTTPWorker()
			}
			// Block until another routine finishes
			jobs <- newJob
	}
}

// Routine that reads continually reads requests from "jobs"
// and quits if it doesn't find any jobs for some time
func asyncHTTPWorker() {
	for {
		select {
			// Get a new job from the queue and process it
			case job := <-jobs:
				res, err := Client.Do(job.req)
				job.c <- JobResult{res, err, job.data}
			// Timeout, kill the routine
			case <-time.After(timeOut):
				atomic.AddInt32(&activeWorkers, -1)
				return
		}
	}
}
