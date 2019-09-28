package apijson

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

func ParseCommentsPage(res *fasthttp.Response, cont *CommentContinuation) (page CommentPage, err error) {
	if res.StatusCode() != fasthttp.StatusOK {
		return page, fmt.Errorf("response status %d", res.StatusCode())
	}

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(res.Body())
	if err != nil {
		return page, err
	}

	var objBlob *fastjson.Value
	if cont.ParentID == "" {
		objBlob = root.Get("response", "continuationContents", "itemSectionContinuation", "contents")
	} else {
		objBlob = root.Get("response", "continuationContents", "commentRepliesContinuation", "contents")
	}
	if objBlob == nil {
		return page, fmt.Errorf("no continuation contents found")
	}
	objArray, err := objBlob.Array()
	if err != nil {
		return page, fmt.Errorf("comments list not found")
	}

	parseBegin := time.Now()
	for _, obj := range objArray {
		var c data.Comment
		c.CrawledAt = parseBegin.Unix()
		var cErr error
		if cont.ParentID == "" {
			cErr = parseThread(&c, obj)
		} else {
			cErr = parseComment(&c, obj)
		}
		if cErr != nil {
			page.CommentParseErrs = append(page.CommentParseErrs, cErr)
			continue
		}
		page.Comments = append(page.Comments, c)
	}
	page.Continuation = cont

	// Parse continuation
	continuationSection := root.Get("response", "continuationContents", "itemSectionContinuation")
	continuationBlob := string(continuationSection.GetStringBytes(
		"continuations", "0", "nextContinuationData", "continuation"))
	if continuationBlob != "" {
		cont.Token = continuationBlob
		page.MoreComments = true
	}
	// Other continuation streams
	otherContObjs := continuationSection.GetArray("header", "commentsHeaderRenderer",
		"sortMenu", "sortFilterSubMenuRenderer", "subMenuItems")
	for _, contObj := range otherContObjs {
		token := string(contObj.GetStringBytes("continuation", "reloadContinuationData", "continuation"))
		title := contObj.GetStringBytes("title")
		otherCont := &CommentContinuation{
			ParentID: cont.ParentID,
			Cookie:   cont.Cookie,
			Token:    token,
			XSRF:     cont.XSRF,
		}
		switch {
		case bytes.Equal(title, []byte("Top comments")):
			page.TopComments = otherCont
		case bytes.Equal(title, []byte("Newest first")):
			page.NewComments = otherCont
		}
	}
	return page, nil
}

func parseThread(c *data.Comment, commentRoot *fastjson.Value) error {
	threadRenderer := commentRoot.Get("commentThreadRenderer")
	//c.RenderingPriority = string(commentRoot.GetStringBytes("renderingPriority"))
	obj := threadRenderer.Get("comment")
	err := parseComment(c, obj)
	if err != nil {
		return err
	}

	// Continuation token
	cont := string(threadRenderer.GetStringBytes(
		"replies", "commentRepliesRenderer",
		"continuations", "0", "nextContinuationData", "continuation"))
	if cont != "" {
		c.Internal = &commentData{
			continuation: CommentContinuation{
				ParentID: c.ID,
				Token:    cont,
			},
		}
	}

	return nil
}

func parseComment(c *data.Comment, obj *fastjson.Value) error {
	obj = obj.Get("commentRenderer")
	c.ID = string(obj.GetStringBytes("commentId"))
	if c.ID == "" {
		return fmt.Errorf("no ID")
	}
	cidDot := strings.IndexByte(c.ID, '.')
	if cidDot >= 0 {
		c.ParentID = c.ID[:cidDot]
	}
	c.LikeCount = obj.GetUint64("likeCount")
	c.ReplyCount = obj.GetUint64("replyCount")

	// Author
	c.Author = string(obj.GetStringBytes("authorText", "simpleText"))
	authorEndPoint := obj.Get("authorEndpoint", "browseEndpoint")
	c.AuthorID = string(authorEndPoint.GetStringBytes("browseId"))

	// Comment text
	/*var builder strings.Builder
	contentText := obj.GetArray("contentText", "runs")
	for _, run := range contentText {
		text := string(run.GetStringBytes("text"))
		if strings.TrimSpace(text) != "" {
			builder.WriteString(text)
		}
		builder.WriteByte('\n')
	}
	c.Text = builder.String()*/
	contentText := obj.Get("contentText", "runs")
	if contentText == nil {
		return fmt.Errorf("no text found")
	}
	c.Content = contentText.MarshalTo(nil)

	// Published Time & video ID
	publishedTimeRun := obj.Get("publishedTimeText", "runs", "0")
	c.CreatedText = string(publishedTimeRun.GetStringBytes("text"))
	c.VideoID = string(publishedTimeRun.GetStringBytes("navigationEndpoint", "watchEndpoint", "videoId"))
	if c.VideoID == "" {
		return fmt.Errorf("failed to find video ID")
	}
	if strings.Contains(c.CreatedText, "edited") {
		c.Edited = true
	}
	var err error
	c.CreatedBefore, c.CreatedAfter, err = parsePublishedTime(c.CreatedText)
	if err != nil {
		return err
	}
	return nil
}

type commentData struct {
	continuation CommentContinuation
}

func CommentRepliesContinuation(c *data.Comment, prev *CommentContinuation) *CommentContinuation {
	commentData, ok := c.Internal.(*commentData)
	if !ok {
		return nil
	}
	return &CommentContinuation{
		ParentID: commentData.continuation.ParentID,
		Cookie:   prev.Cookie,
		Token:    commentData.continuation.Token,
		XSRF:     prev.XSRF,
	}
}

// parses "8 hours ago" to a time interval
func parsePublishedTime(str string) (after int64, before int64, err error) {
	parts := strings.SplitN(str, " ", 4)
	if len(parts) < 3 || parts[2] != "ago" {
		err = fmt.Errorf("couldn't parse time range: %s", str)
		return
	}
	// Parse amount
	amount, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		err = fmt.Errorf("couldn't parse time amount: %s", str)
		return
	}
	// Parse precision
	var precision time.Duration
	switch {
	case strings.HasPrefix(parts[1], "second"):
		precision = time.Second
	case strings.HasPrefix(parts[1], "minute"):
		precision = time.Minute
	case strings.HasPrefix(parts[1], "hour"):
		precision = time.Hour
	case strings.HasPrefix(parts[1], "day"):
		precision = 24 * time.Hour
	case strings.HasPrefix(parts[1], "week"):
		precision = 7 * 24 * time.Hour
	// TODO More accurate interval
	case strings.HasPrefix(parts[1], "month"):
		precision = 30 * 7 * 24 * time.Hour
	case strings.HasPrefix(parts[1], "year"):
		precision = 365 * 7 * 24 * time.Hour
	default:
		err = fmt.Errorf("unknown precision: %s", precision)
		return
	}
	// Build time interval
	beforeT := time.Now().Add(-time.Duration(amount) * precision)
	afterT := beforeT.Add(-precision)
	return beforeT.Unix(), afterT.Unix(), nil
}


type CommentPage struct {
	Comments []data.Comment
	CommentParseErrs []error
	MoreComments bool
	Continuation *CommentContinuation
	TopComments *CommentContinuation
	NewComments *CommentContinuation
}

type CommentContinuation struct {
	ParentID string
	Cookie string
	Token  string
	XSRF   string
}

