package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Scheduler struct {
	Ctx     context.Context
	server  http.Server
	out     chan<- []byte
	jobLock sync.RWMutex
	jobs    map[*Job]bool
}

type Job struct {
	Started time.Time `json:"started"`
	VideoID string    `json:"videoID"`
	Sort    string    `json:"mode"`
	Pages   int64     `json:"pages"`
	Items   int64     `json:"items"`
}

func (s *Scheduler) Init(ctx context.Context, out chan<- []byte) error {
	httpListen := os.Getenv("HTTP_LISTEN")
	if httpListen == "" {
		return fmt.Errorf("$HTTP_LISTEN not set")
	}

	engine := gin.Default()
	engine.GET("/", func(context *gin.Context) {
		s.jobLock.RLock()
		jobs := make([]*Job, len(s.jobs))
		i := 0
		for job := range s.jobs {
			jobs[i] = job
			i++
		}
		s.jobLock.RUnlock()
		context.JSON(http.StatusOK, jobs)
	})
	engine.POST("/", func(context *gin.Context) {
		context.SetAccepted("application/json")
		var requestData struct {
			VideoID string `json:"videoID"`
			Sort    string `json:"sort"`
		}
		err := context.BindJSON(&requestData)
		if err != nil {
			_ = context.Error(err)
			return
		}
		job := &Job{
			Started: time.Now(),
			VideoID: requestData.VideoID,
			Sort:    requestData.Sort,
		}
		context.JSON(http.StatusCreated, job)
		s.startWorker(out, job)
	})

	s.Ctx = ctx
	s.server = http.Server{
		Addr:    httpListen,
		Handler: engine,
	}
	s.out = out
	s.jobs = make(map[*Job]bool)
	return nil
}

func (s *Scheduler) Run() {
	err := s.server.ListenAndServe()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to start HTTP server")
	}
}

func (s *Scheduler) Close() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_ = s.server.Shutdown(ctx)
}
