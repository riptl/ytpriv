package common

import "net/http"

type JobResult struct {
	Res *http.Response
	Err error
	ReqData interface{} // job.data
}

type job struct {
	req *http.Request
	c chan JobResult
	data interface{}
}

var jobs = make(chan job)

func InitAsyncHTTP(nWorkers uint) {
	for i := uint(0); i < nWorkers; i++ {
		go asyncHTTPWorker()
	}
}

func DoAsyncHTTP(r *http.Request, c chan JobResult, data interface{}) {
	jobs <- job{r, c, data}
}

func asyncHTTPWorker() {
	for {
		job := <-jobs
		res, err := Client.Do(job.req)
		job.c <- JobResult{res, err, job.data}
	}
}
