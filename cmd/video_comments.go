package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

var continuationLimitReached = fmt.Errorf("continuation limit reached")

var videoCommentsCmd = cobra.Command{
	Use:   "comments [video...]",
	Short: "Scrape comments of videos",
	Args:  cobra.ArbitraryArgs,
	Run:   cmdFunc(doVideoComments),
}

func init() {
	f := videoCommentsCmd.Flags()
	f.Duration("slow-start", time.Second, "Time to wait between opening connections")
	f.String("sort", "new", "Comment sort order (new, top)")
}

func doVideoComments(c *cobra.Command, args []string) error {
	flags := c.Flags()
	sortOrder, err := flags.GetString("sort")
	if err != nil {
		panic(err)
	}
	var top bool
	switch sortOrder {
	case "new":
		top = false
	case "top":
		top = true
	default:
		return fmt.Errorf("unknown sort order: %s", sortOrder)
	}

	start := time.Now()
	defer func() {
		logrus.Infof("Finished after %s", time.Since(start).String())
	}()

	// Create argument channels
	jobs := make(chan string)
	go stdinOrArgs(jobs, args)
	videoIDs := make(chan string)
	go func() {
		defer close(videoIDs)
		for job := range jobs {
			videoID, err := api.GetVideoID(job)
			if err != nil {
				logrus.Error(err)
				continue
			}
			videoIDs <- videoID
		}
	}()

	// Dump comments
	comments := make(chan data.Comment)
	startDelay, err := flags.GetDuration("slow-start")
	if err != nil {
		panic(err)
	}
	go videoCommentDumpScheduler(comments, videoIDs, startDelay, top)
	enc := json.NewEncoder(os.Stdout)
	commentCount := int64(0)
	for comment := range comments {
		if err := enc.Encode(&comment); err != nil {
			return err
		}
		commentCount++
	}
	logrus.WithField("count", commentCount).Info("Success")

	return nil
}

func videoCommentDumpScheduler(comments chan<- data.Comment, videoIDs <-chan string, startDelay time.Duration, top bool) {
	defer close(comments)
	var wg sync.WaitGroup
	wg.Add(int(net.MaxWorkers))
	for i := uint(0); i < net.MaxWorkers; i++ {
		go func() {
			defer wg.Done()
			for videoID := range videoIDs {
				logrus.WithField("video_id", videoID).
					Info("Start video")
				err := streamVideoComments(comments, videoID, top)
				if err != nil {
					logrus.WithError(err).Errorf("Failed to dump comments of video %s", videoID)
				}
			}
		}()
		time.Sleep(startDelay)
	}
	wg.Wait()
}

func streamVideoComments(comments chan<- data.Comment, videoID string, top bool) error {
	videoID, err := api.GetVideoID(videoID)
	if err != nil {
		return err
	}

	vid, err := simpleGetVideo(videoID)
	if err != nil {
		return err
	}

	cont := api.InitialCommentContinuation(vid)
	if cont == nil {
		return fmt.Errorf("failed to request comments")
	}

	if top {
		streamTopComments(comments, cont)
	} else {
		streamNewComments(comments, cont)
	}

	return nil
}

func simpleGetVideo(videoID string) (v *data.Video, err error) {
	videoReq := api.GrabVideo(videoID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = net.Client.Do(videoReq, res)
	if err != nil {
		return nil, err
	}

	v = new(data.Video)
	v.ID = videoID
	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		err = api.ParseVideo(v, res)
		if err == api.ErrRateLimit {
			logrus.WithField("video_id", videoID).Warnf("Rate-Limited (%d/%d)", i+1, maxRetries)
			time.Sleep(time.Second)
			continue
		}
		if err == nil {
			return v, nil
		} else {
			return nil, err
		}
	}
	return nil, api.ErrRateLimit
}

func streamComments(comments chan<- data.Comment, cont *api.CommentContinuation, i int) {
	var j int
	var err error
	for {
		var page api.CommentPage
		page, err = nextCommentPage(cont, i, j)
		if err != nil {
			break
		}
		if cont.ParentID != "" {
			j++
		} else {
			i++
		}

		for _, comment := range page.Comments {
			subCont := api.CommentRepliesContinuation(&comment, cont)
			if subCont != nil {
				streamComments(comments, subCont, i)
			}
			comments <- comment
		}

		if !page.MoreComments {
			break
		}
	}
	if err == continuationLimitReached {
		logrus.Warn("Continuation limit reached")
	} else if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
	}
}

func streamNewComments(comments chan<- data.Comment, cont *api.CommentContinuation) {
	var err error
	var page api.CommentPage
	page, err = nextCommentPage(cont, -1, 0)
	if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
		return
	}
	if page.NewComments == nil {
		return
	}
	*cont = *page.NewComments
	streamComments(comments, cont, 0)
}

func streamTopComments(comments chan<- data.Comment, cont *api.CommentContinuation) {
	var err error
	var page api.CommentPage
	page, err = nextCommentPage(cont, -1, 0)
	if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
		return
	}
	if page.NewComments == nil {
		return
	}
	*cont = *page.TopComments
	streamComments(comments, cont, 0)
}

func nextCommentPage(cont *api.CommentContinuation, i int, j int) (page api.CommentPage, err error) {
	req := api.GrabCommentPage(cont)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	err = net.Client.Do(req, res)
	if err != nil {
		return page, err
	}
	switch res.StatusCode() {
	case fasthttp.StatusRequestEntityTooLarge,
		fasthttp.StatusRequestURITooLong:
		return page, continuationLimitReached
	}

	page, err = api.ParseCommentsPage(res, cont)
	if err != nil {
		return page, err
	}
	for _, cErr := range page.CommentParseErrs {
		logrus.WithError(cErr).Error("Failed to parse comment")
	}

	if cont.ParentID == "" {
		logrus.WithFields(logrus.Fields{
			"video_id": cont.VideoID,
			"index":    i,
		}).Infof("Page")
	} else {
		logrus.WithFields(logrus.Fields{
			"video_id":  cont.VideoID,
			"index":     i,
			"sub_index": j,
			"parent_id": cont.ParentID,
		}).Infof("Sub page")
	}
	return page, nil
}
