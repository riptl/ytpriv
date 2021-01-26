package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yt "github.com/terorie/ytpriv"
	"github.com/terorie/ytpriv/types"
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

	videoCmd.AddCommand(&videoCommentsCmd)
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
			videoID, err := yt.ExtractVideoID(job)
			if err != nil {
				logrus.Error(err)
				continue
			}
			videoIDs <- videoID
		}
	}()

	// Dump comments
	comments := make(chan types.Comment)
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

func videoCommentDumpScheduler(comments chan<- types.Comment, videoIDs <-chan string, startDelay time.Duration, top bool) {
	defer close(comments)
	var wg sync.WaitGroup
	wg.Add(int(maxWorkers))
	for i := uint(0); i < maxWorkers; i++ {
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

func streamVideoComments(comments chan<- types.Comment, videoID string, top bool) error {
	videoID, err := yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}
	vid, err := client.RequestVideo(videoID).Do()
	if err != nil {
		return err
	}
	cont := yt.InitialCommentContinuation(vid)
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

func streamComments(comments chan<- types.Comment, cont *types.CommentContinuation, i int) {
	var j int
	var err error
	for {
		var page types.CommentPage
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
			subCont := yt.CommentRepliesContinuation(&comment, cont)
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

func streamNewComments(comments chan<- types.Comment, cont *types.CommentContinuation) {
	var err error
	var page types.CommentPage
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

func streamTopComments(comments chan<- types.Comment, cont *types.CommentContinuation) {
	var err error
	var page types.CommentPage
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

func nextCommentPage(cont *types.CommentContinuation, i int, j int) (page types.CommentPage, err error) {
	page, err = client.RequestCommentPage(cont).Do()
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
