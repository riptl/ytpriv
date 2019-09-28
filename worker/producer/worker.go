package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apijson"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var continuationLimitReached = fmt.Errorf("continuation limit reached")

func (s *Scheduler) startWorker(out chan<- []byte, job *Job) {
	s.jobLock.Lock()
	s.jobs[job] = true
	s.jobLock.Unlock()

	go func() {
		err := streamComments(out, job)
		if err != nil {
			logrus.WithField("video", job.VideoID).WithError(err).
				Error("Failed to stream comments of video")
		}
		s.jobLock.Lock()
		delete(s.jobs, job)
		s.jobLock.Unlock()
	}()
}

func streamComments(out chan<- []byte, job *Job) error {
	videoID, err := api.GetVideoID(job.VideoID)
	if err != nil { return err }

	vid, err := simpleGetVideo(videoID)
	if err != nil { return err }

	cont := apijson.InitialCommentContinuation(vid)
	if cont == nil {
		return fmt.Errorf("failed to request comments")
	}

	j := videoCommentsJob{
		Job: job,
		log: logrus.WithField("video", job.VideoID),
		comments: make(chan data.Comment),
	}
	var runFunc func()
	switch j.Job.Sort {
	case "top":
		runFunc = func() { j.streamComments(cont) }
	case "age":
		runFunc = func() { j.streamNewComments(cont) }
	case "live":
		runFunc = func() { j.streamLiveComments(cont) }
	default:
		return fmt.Errorf("unknown sort order: %s", j.Sort)
	}
	j.wg.Add(1)
	go func() {
		defer j.wg.Done()
		runFunc()
	}()
	go func() {
		j.wg.Wait()
		close(j.comments)
	}()

	for comment := range j.comments {
		commentBuf, err := json.Marshal(&comment)
		if err != nil { panic(err) }
		out <- commentBuf
	}
	return nil
}

// TODO Copied

type videoCommentsJob struct {
	*Job
	log *logrus.Entry
	wg sync.WaitGroup
	comments chan data.Comment
}

func simpleGetVideo(videoID string) (v *data.Video, err error) {
	videoReq := apijson.GrabVideo(videoID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = backoff.Retry(func() error {
		err = net.Client.Do(videoReq, res)
		if err == fasthttp.ErrNoFreeConns {
			logrus.WithError(err).Warn("No free conns, throttling")
			return err
		} else if err != nil {
			return backoff.Permanent(err)
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return nil, err
	}

	v = new(data.Video)
	v.ID = videoID
	err = apijson.ParseVideo(v, res)
	if err != nil { return nil, err }

	return
}

func (j *videoCommentsJob) streamComments(cont *apijson.CommentContinuation) {
	var err error
	for i := 0; true; i++ {
		var page apijson.CommentPage
		page, err = j.nextCommentPage(cont, i)
		if err != nil {
			break
		}

		for _, comment := range page.Comments {
			subCont := apijson.CommentRepliesContinuation(&comment, cont)
			if subCont != nil {
				j.wg.Add(1)
				go func() {
					defer j.wg.Done()
					j.streamComments(subCont)
				}()
			}
			j.comments <- comment
		}

		if !page.MoreComments {
			break
		}
	}
	if err == continuationLimitReached {
		j.log.Warn("Continuation limit reached")
	} else if err != nil {
		j.log.WithError(err).Error("Comment stream aborted")
	} else if cont.ParentID == "" {
		j.log.Info("Finished comment stream")
	} else {
		j.log.Infof("Finished sub comment stream (%s)", cont.ParentID)
	}
}

func (j *videoCommentsJob) streamNewComments(cont *apijson.CommentContinuation) {
	var err error
	var page apijson.CommentPage
	page, err = j.nextCommentPage(cont, -1)
	if err != nil {
		j.log.WithError(err).Error("Comment stream aborted")
		return
	}
	*cont = *page.NewComments

	j.streamComments(cont)
}

func (j *videoCommentsJob) streamLiveComments(cont *apijson.CommentContinuation) {
	// TODO Basic deduplication

	var err error
	var page apijson.CommentPage
	page, err = j.nextCommentPage(cont, -1)
	if err != nil {
		j.log.WithError(err).Error("Comment stream aborted")
		return
	}
	*cont = *page.NewComments

	for i := 0; true; i++ {
		page, err = j.nextCommentPage(cont, i)
		if err != nil {
			break
		}
		*cont = *page.NewComments
		for _, comment := range page.Comments {
			subCont := apijson.CommentRepliesContinuation(&comment, cont)
			if subCont != nil {
				j.wg.Add(1)
				go func() {
					defer j.wg.Done()
					j.streamComments(subCont)
				}()
			}
			j.comments <- comment
		}
	}
	if err != nil {
		j.log.WithError(err).Error("Comment stream aborted")
	}
}

func (j *videoCommentsJob) nextCommentPage(cont *apijson.CommentContinuation, i int) (page apijson.CommentPage, err error) {
	req := apijson.GrabCommentPage(cont)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	err = backoff.Retry(func() error {
		err = net.Client.Do(req, res)
		if err == fasthttp.ErrNoFreeConns {
			logrus.WithError(err).Warn("No free conns, throttling")
			return err
		} else if err != nil {
			return backoff.Permanent(err)
		}
		return nil
	}, backoff.NewExponentialBackOff())
	if err != nil {
		return page, err
	}
	switch res.StatusCode() {
	case fasthttp.StatusRequestEntityTooLarge,
		fasthttp.StatusRequestURITooLong:
		return page, continuationLimitReached
	}

	page, err = apijson.ParseCommentsPage(res, cont)
	if err != nil { return page, err }
	for _, cErr := range page.CommentParseErrs {
		j.log.WithError(cErr).Error("Failed to parse comment")
	}

	if cont.ParentID == "" {
		j.log.Infof("Got page #%02d", i)
	} else {
		j.log.Infof("Got sub comment page (%s) #%02d", cont.ParentID, i)
	}
	atomic.AddInt64(&j.Pages, 1)
	atomic.AddInt64(&j.Items, int64(len(page.Comments)))
	return page, nil
}

