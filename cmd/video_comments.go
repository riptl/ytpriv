package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apijson"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var continuationLimitReached = fmt.Errorf("continuation limit reached")

var videoCommentsCmd = cobra.Command{
	Use: "comments <video>",
	Short: "Scrape comments on a video",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doVideoComments),
}

func init() {
	flags := videoCommentsCmd.Flags()
	flags.StringP("sort", "s", "top", `"top": sort by likes, "age": new to old, "live": newest page (infinite)`)
}

type videoCommentsJob struct {
	wg sync.WaitGroup
	comments chan data.Comment
}

func doVideoComments(c *cobra.Command, args []string) error {
	if apis.Main != &apis.JsonAPI {
		return fmt.Errorf("only JSON API supported")
	}

	videoID := args[0]

	videoID, err := api.GetVideoID(videoID)
	if err != nil { return err }

	vid, err := simpleGetVideo(videoID)
	if err != nil { return err }

	cont := apijson.InitialCommentContinuation(vid)
	if cont == nil {
		return fmt.Errorf("failed to request comments")
	}

	j := videoCommentsJob{
		comments: make(chan data.Comment),
	}
	j.wg.Add(1)

	sortFlag, err := c.Flags().GetString("sort")
	if err != nil {
		return err
	}
	switch sortFlag {
	case "top":
		go j.streamComments(cont)
	case "age":
		go j.streamNewComments(cont)
	case "live":
		go j.streamLiveComments(cont)
	default:
		return fmt.Errorf("unknown sort order: %s", sortFlag)
	}
	go func() {
		j.wg.Wait()
		close(j.comments)
	}()

	enc := json.NewEncoder(os.Stdout)
	for comment := range j.comments {
		err := enc.Encode(&comment)
		if err != nil { panic(err) }
	}

	return nil
}

func simpleGetVideo(videoID string) (v *data.Video, err error) {
	videoReq := apis.Main.GrabVideo(videoID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = net.Client.Do(videoReq, res)
	if err != nil { return nil, err }

	v = new(data.Video)
	v.ID = videoID
	err = apis.Main.ParseVideo(v, res)
	if err != nil { return nil, err }

	return
}

func (j *videoCommentsJob) streamComments(cont *apijson.CommentContinuation) {
	defer j.wg.Done()
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
				go j.streamComments(subCont)
			}
			j.comments <- comment
		}

		if !page.MoreComments {
			break
		}
	}
	if err == continuationLimitReached {
		logrus.Warn("Continuation limit reached")
	} else if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
	} else {
		logrus.Info("Finished comment stream")
	}
}

func (j *videoCommentsJob) streamNewComments(cont *apijson.CommentContinuation) {
	defer j.wg.Done()
	var err error
	var page apijson.CommentPage
	page, err = j.nextCommentPage(cont, -1)
	if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
		return
	}
	*cont = *page.NewComments

	j.streamComments(cont)
}

func (j *videoCommentsJob) streamLiveComments(cont *apijson.CommentContinuation) {
	// TODO Basic deduplication

	defer j.wg.Done()
	var err error
	var page apijson.CommentPage
	page, err = j.nextCommentPage(cont, -1)
	if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
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
				go j.streamComments(subCont)
			}
			j.comments <- comment
		}
	}
	if err != nil {
		logrus.WithError(err).Error("Comment stream aborted")
	}
}

func (j *videoCommentsJob) nextCommentPage(cont *apijson.CommentContinuation, i int) (page apijson.CommentPage, err error) {
	req := apijson.GrabCommentPage(cont)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	err = net.Client.Do(req, res)
	if err != nil { return page, err }
	switch res.StatusCode() {
	case fasthttp.StatusRequestEntityTooLarge,
		fasthttp.StatusRequestURITooLong:
		return page, continuationLimitReached
	}

	page, err = apijson.ParseCommentsPage(res, cont)
	if err != nil { return page, err }
	for _, cErr := range page.CommentParseErrs {
		logrus.WithError(cErr).Error("Failed to parse comment")
	}

	if cont.ParentID == "" {
		logrus.Infof("Got page #%02d\n", i)
	} else {
		logrus.Infof("Got sub comment page (%s) #%02d\n", cont.ParentID, i)
	}
	return page, nil
}
