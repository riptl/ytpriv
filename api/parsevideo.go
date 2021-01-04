package api

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/terorie/ytwrk/data"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

var matchThumbUrl = regexp.MustCompile("^.+/hqdefault\\.jpg")

var unexpectedType = errors.New("unexpected type")
var ErrRateLimit = errors.New("reCAPTCHA triggered")

func ParseVideo(v *data.Video, res *fasthttp.Response) error {
	if res.StatusCode() != fasthttp.StatusOK {
		return fmt.Errorf("response status %d", res.StatusCode())
	}
	contentType := res.Header.ContentType()
	switch {
	case bytes.HasPrefix(contentType, []byte("application/json")):
		break
	case bytes.HasPrefix(contentType, []byte("text/html")):
		return ErrRateLimit
	}

	return ParseVideoBody(v, res.Body(), res)
}

// res is optional
func ParseVideoBody(v *data.Video, buf []byte, res *fasthttp.Response) error {
	var internal *videoData
	if res != nil {
		internal = new(videoData)
		v.Internal = internal
	}

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(buf)
	if err != nil {
		return err
	}

	rootArray := root.GetArray()
	if rootArray == nil {
		return unexpectedType
	}

	// Get interesting objects
	var pageResponse *fastjson.Value
	var playerResponse *fastjson.Value
	var playerArgs *fastjson.Value
	var xsrfToken string
	for _, sub := range rootArray {
		if playerResponse == nil {
			playerResponse = sub.Get("playerResponse")
		}
		if pageResponse == nil {
			pageResponse = sub.Get("response")
		}
		if xsrfToken == "" {
			xsrfToken = string(sub.GetStringBytes("xsrf_token"))
		}
		if playerArgs == nil {
			playerArgs = sub.Get("player", "args")
		}
	}

	if playerResponse == nil {
		return errors.New("no video details")
	}
	if pageResponse == nil {
		return errors.New("no page response")
	}

	// Playability status
	playability := playerResponse.Get("playabilityStatus")

	// Playable at all?
	playabilityStatus := string(playability.GetStringBytes("status"))
	switch playabilityStatus {
	case "ERROR":
		return VideoUnavailable
	case "LOGIN_REQUIRED":
		v.FamilyFriendly = false
	default:
		v.FamilyFriendly = true
	}

	// Playable in embed?
	playableInEmbedValue := playability.Get("playableInEmbed")
	if playableInEmbedValue.Exists() {
		playableInEmbed, _ := playableInEmbedValue.Bool()
		v.NoEmbed = !playableInEmbed
	}

	if err := parseVideoDetails(v, playerResponse.Get("videoDetails")); err != nil {
		return err
	}
	if err := parseMicroformat(v, playerResponse.Get("microformat")); err != nil {
		return err
	}

	watchNextResults := pageResponse.Get("contents", "twoColumnWatchNextResults")

	// Parse video infos
	watchNextContents := watchNextResults.GetArray("results", "results", "contents")
	if err := parseVideoInfo(v, watchNextContents); err != nil {
		return err
	}

	if res != nil {
		var visitorInfo, ysc, cookie string
		var ok bool
		visitorInfo, ok = parseSetCookie(res, "VISITOR_INFO1_LIVE")
		if !ok {
			goto cookieFailed
		}
		ysc, ok = parseSetCookie(res, "YSC")
		if !ok {
			goto cookieFailed
		}
		cookie = visitorInfo + "; " + ysc
		parseCommentToken(internal, watchNextContents, v.ID, cookie, xsrfToken)
	}

cookieFailed:

	// TODO secondaryResults

	// Get related vids
	v.RelatedVideos = parseVideoRelated(watchNextResults)

	// Parse player args
	if playerArgs != nil {
		if err := parsePlayerArgs(v, playerArgs); err != nil {
			return err
		}
	}

	// Get captions
	captionTracks := playerResponse.GetArray("captions", "playerCaptionsTracklistRenderer", "captionTracks")
	for _, track := range captionTracks {
		v.Captions = append(v.Captions, data.Caption{
			VssID:        string(track.GetStringBytes("vssId")),
			Name:         string(track.GetStringBytes("name", "simpleText")),
			Code:         string(track.GetStringBytes("languageCode")),
			Translatable: track.GetBool("isTranslatable"),
		})
	}

	return nil
}

func parseVideoDetails(v *data.Video, videoDetails *fastjson.Value) error {
	// Get livestream info
	if videoDetails.GetBool("isLive") {
		v.Livestream = new(data.Livestream)

		v.Livestream.OwnerViewing = videoDetails.GetBool("isOwnerViewing")
		v.Livestream.DvrEnabled = videoDetails.GetBool("isLiveDvrEnabled")
		v.Livestream.LowLatency = videoDetails.GetBool("isLowLatencyLiveStream")
		v.Livestream.LiveContent = videoDetails.GetBool("isLiveContent")
	}

	// Get tags
	keywords := videoDetails.GetArray("keywords")
	for _, keywordValue := range keywords {
		keywordBytes, _ := keywordValue.StringBytes()
		if keywordBytes == nil {
			continue
		}

		keyword := string(keywordBytes)
		v.Tags = append(v.Tags, keyword)
	}

	v.ID = string(videoDetails.GetStringBytes("videoId"))
	// Get title
	v.Title = string(videoDetails.GetStringBytes("title"))
	// Get description
	v.Description = string(videoDetails.GetStringBytes("shortDescription"))
	// Get channel ID
	v.UploaderID = string(videoDetails.GetStringBytes("channelId"))
	// Get author
	v.Uploader = string(videoDetails.GetStringBytes("author"))
	// Ratings allowed?
	v.NoRatings = !videoDetails.GetBool("allowRatings")
	// Get view count
	viewCountStr := string(videoDetails.GetStringBytes("viewCount"))
	viewCount64, err := strconv.ParseUint(viewCountStr, 10, 64)
	if err == nil {
		v.Views = viewCount64
	}
	// Get duration
	lengthStr := string(videoDetails.GetStringBytes("lengthSeconds"))
	length64, err := strconv.ParseUint(lengthStr, 10, 64)
	if err == nil {
		v.Duration = length64
	}

	return nil
}

func parseMicroformat(v *data.Video, microformat *fastjson.Value) error {
	renderer := microformat.Get("playerMicroformatRenderer")
	v.Genre = string(renderer.GetStringBytes("category"))
	return nil
}

func parseVideoInfo(v *data.Video, videoInfo []*fastjson.Value) error {
	var primary *fastjson.Value
	var secondary *fastjson.Value

	for _, obj := range videoInfo {
		if primary == nil {
			primary = obj.Get("videoPrimaryInfoRenderer")
		}
		if secondary == nil {
			secondary = obj.Get("videoSecondaryInfoRenderer")
		}
	}
	if primary == nil || secondary == nil {
		return fmt.Errorf("missing video info objects")
	}

	// Check unlisted status (from badges)
	for _, badge := range primary.GetArray("badges") {
		if string(badge.GetStringBytes("metadataBadgeRenderer", "label")) == "Unlisted" {
			v.Visibility = data.VisibilityUnlisted
			break
		}
	}

	// Get like/dislike count
	likeRatioStr := string(primary.GetStringBytes("sentimentBar", "sentimentBarRenderer", "tooltip"))
	likeRatioParts := strings.Split(likeRatioStr, " / ")
	if len(likeRatioParts) == 2 {
		likesStr := likeRatioParts[0]
		dislikesStr := likeRatioParts[1]
		v.Likes, _ = ExtractNumber(likesStr)
		v.Dislikes, _ = ExtractNumber(dislikesStr)
	}

	// Get upload date
	dateText := string(primary.GetStringBytes("dateText", "simpleText"))
	dateText = strings.TrimPrefix(dateText, "Published on ")
	dateText = strings.TrimPrefix(dateText, "Uploaded on ") // Unlisted video
	dateText = strings.TrimPrefix(dateText, "Started streaming on ")
	dateText = strings.TrimPrefix(dateText, "Streamed live on ")

	if date, err := time.Parse("Jan 2, 2006", dateText); err == nil {
		v.Uploaded = date.Unix()
	} else if date, err := time.Parse("01.02.2006", dateText); err == nil {
		v.Uploaded = date.Unix()
	}

	// Get category
	// Find category row
	metaRows := secondary.GetArray("metadataRowContainer", "metadataRowContainerRenderer", "rows")
	for _, obj := range metaRows {
		row := obj.Get("metadataRowRenderer")
		title := string(row.GetStringBytes("title", "simpleText"))
		switch title {
		case "License":
			v.License = string(row.GetStringBytes("contents", "0", "runs", "0", "text"))
		}
	}

	return nil
}

func parsePlayerArgs(v *data.Video, args *fastjson.Value) error {
	newPlayerResJSON := args.GetStringBytes("player_response")
	if len(newPlayerResJSON) == 0 {
		return nil
	}
	res, err := fastjson.ParseBytes(newPlayerResJSON)
	if err != nil {
		return err
	}
	streamingData := res.Get("streamingData")
	for _, format := range streamingData.GetArray("formats") {
		if itag := format.GetInt("itag"); itag != 0 {
			v.Formats = append(v.Formats, strconv.Itoa(itag))
		}
	}
	for _, format := range streamingData.GetArray("adaptiveFormats") {
		if itag := format.GetInt("itag"); itag != 0 {
			v.Formats = append(v.Formats, strconv.Itoa(itag))
		}
	}
	return nil
}

func parseVideoRelated(watchNextResults *fastjson.Value) []data.RelatedVideo {
	var related []data.RelatedVideo
	results := watchNextResults.GetArray("secondaryResults", "secondaryResults", "results")
	for _, obj := range results {
		renderer := obj.Get("compactVideoRenderer")

		videoId := string(renderer.GetStringBytes("videoId"))
		uploaderID := getVideoRendererUploader(renderer)

		var vid data.RelatedVideo
		vid.ID = videoId

		if uploaderID != "" {
			vid.UploaderID = uploaderID
		}

		if videoId != "" {
			related = append(related, vid)
		}
	}
	return related
}

func parseCommentToken(data *videoData, contentList []*fastjson.Value, videoID string, cookie string, xsrfToken string) {
	for _, content := range contentList {
		sectionIdentifier := string(content.GetStringBytes("itemSectionRenderer", "sectionIdentifier"))
		if sectionIdentifier != "comment-item-section" {
			continue
		}
		continuation := string(content.GetStringBytes("itemSectionRenderer", "continuations", "0", "nextContinuationData", "continuation"))
		if continuation == "" {
			continue
		}
		data.continuation = &CommentContinuation{
			VideoID: videoID,
			Cookie:  cookie,
			Token:   continuation,
			XSRF:    xsrfToken,
		}
		return
	}
}

func getVideoRendererUploader(renderer *fastjson.Value) string {
	channelRuns := renderer.GetArray("longBylineText", "runs")
	for _, run := range channelRuns {
		browseID := string(run.GetStringBytes(
			"navigationEndpoint", "browseEndpoint", "browseId"))
		if browseID != "" {
			return browseID
		}
	}
	return ""
}

func InitialCommentContinuation(v *data.Video) *CommentContinuation {
	videoData, ok := v.Internal.(*videoData)
	if !ok {
		return nil
	}
	return videoData.continuation
}

func parseSetCookie(res *fasthttp.Response, field string) (cookie string, ok bool) {
	cookieBytes := res.Header.PeekCookie(field)
	i := bytes.IndexByte(cookieBytes, ';')
	if i < 0 {
		return "", false
	}
	return string(cookieBytes[:i]), true
}

type videoData struct {
	continuation *CommentContinuation
}
