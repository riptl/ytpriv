package apijson

import (
	"github.com/valyala/fastjson"
	"github.com/terorie/yt-mango/data"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
	"regexp"
	"github.com/terorie/yt-mango/api"
)

var matchThumbUrl = regexp.MustCompile("^.+/hqdefault\\.jpg")

var unexpectedType = errors.New("unexpected type")

func ParseVideo(v *data.Video, res *http.Response) error {
	defer res.Body.Close()

	// Download response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return err }

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(body)
	if err != nil { return err }

	rootArray := root.GetArray()
	if rootArray == nil { return unexpectedType }

	// Get interesting objects
	var pageResponse *fastjson.Value
	var playerResponse *fastjson.Value
	var playerArgs *fastjson.Value
	for _, sub := range rootArray {
		if playerResponse == nil {
			playerResponse = sub.Get("playerResponse")
			pageResponse = sub.Get("response")
			v.URL = string(sub.GetStringBytes("url"))
		}
		if playerArgs == nil {
			playerArgs = sub.Get("player", "args")
		}
	}

	if v.URL != "" { v.URL = "https://www.youtube.com" + v.URL }
	if playerResponse == nil { return errors.New("no video details") }

	// Playability status
	playability := playerResponse.Get("playabilityStatus")

	// Playable at all?
	playabilityStatus := string(playability.GetStringBytes("status"))
	switch playabilityStatus {
	case "ERROR":
		return api.VideoUnavailable
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

	if err := parseVideoDetails(v, playerResponse.Get("videoDetails"));
		err != nil { return err }

	watchNextResults := pageResponse.Get("contents", "twoColumnWatchNextResults")

	// Parse video infos
	watchNextContents := watchNextResults.GetArray("results", "results", "contents")
	if err := parseVideoInfo(v, watchNextContents);
		err != nil { return err }

	// Get related vids
	v.Related = parseVideoRelated(watchNextResults)

	// Parse player args
	if playerArgs != nil {
		if err := parsePlayerArgs(v, playerArgs);
			err != nil { return err }
	}

	return nil
}

func parseVideoDetails(v *data.Video, videoDetails *fastjson.Value) error {
	// Get tags
	keywords := videoDetails.GetArray("keywords")
	for _, keywordValue := range keywords {
		keywordBytes, _ := keywordValue.StringBytes()
		if keywordBytes == nil { continue }

		keyword := string(keywordBytes)
		v.Tags = append(v.Tags, keyword)
	}

	// Get title
	v.Title = string(videoDetails.GetStringBytes("title"))
	// Get description
	v.Description = string(videoDetails.GetStringBytes("shortDescription"))
	// Get channel ID
	v.UploaderID = string(videoDetails.GetStringBytes("channelId"))
	if v.UploaderID != "" {
		v.UploaderURL = "https://www.youtube.com/channel/" + v.UploaderID
	}
	// Get author
	v.Uploader = string(videoDetails.GetStringBytes("author"))
	// Ratings allowed?
	v.NoRatings = !videoDetails.GetBool("allowRatings")
	// Get view count
	viewCountStr := string(videoDetails.GetStringBytes("viewCount"))
	viewCount64, err := strconv.ParseUint(viewCountStr, 10, 64)
	if err == nil { v.Views = viewCount64 }
	// Get duration
	lengthStr := string(videoDetails.GetStringBytes("lengthSeconds"))
	length64, err := strconv.ParseUint(lengthStr, 10, 64)
	if err == nil { v.Duration = length64 }

	// Get thumbnail URL
	thumbUrl := string(videoDetails.GetStringBytes("thumbnail", "thumbnails", "0", "url"))
	if thumbUrl != "" {
		indices := matchThumbUrl.FindStringIndex(thumbUrl)
		if len(indices) == 2 {
			v.Thumbnail = thumbUrl[indices[0]:indices[1]]
		}
	}

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
		v.Likes, _ = api.ExtractNumber(likesStr)
		v.Dislikes, _ = api.ExtractNumber(dislikesStr)
	}

	// Get upload date
	dateText := string(secondary.GetStringBytes("dateText", "simpleText"))
	dateText = strings.TrimPrefix(dateText, "Published on ")
	dateText = strings.TrimPrefix(dateText, "Uploaded on ") // Unlisted video

	date, err := time.Parse("Jan _2, 2006", dateText)
	if err == nil { v.UploadDate = date }

	// Get category
	// Find category row
	metaRows := secondary.GetArray("metadataRowContainer", "metadataRowContainerRenderer", "rows")
	for _, obj := range metaRows {
		row := obj.Get("metadataRowRenderer")
		title := string(row.GetStringBytes("title", "simpleText"))
		switch title {
		case "Category":
			v.Genre = string(row.GetStringBytes("contents", "0", "runs", "0", "text"))
		case "License":
			v.License = string(row.GetStringBytes("contents", "0", "runs", "0", "text"))
		}
	}

	return nil
}

func parsePlayerArgs(v *data.Video, args *fastjson.Value) error {
	fmts := string(args.GetStringBytes("fmt_list"))
	fmtList, err := api.ParseFormatList(fmts)
	if err != nil { return err }
	v.Formats = fmtList
	return nil
}

func parseVideoRelated(watchNextResults *fastjson.Value) []string {
	var related []string
	results := watchNextResults.GetArray("secondaryResults", "secondaryResults", "results")
	for _, obj := range results {
		videoId := string(obj.GetStringBytes("compactVideoRenderer", "videoId"))
		if videoId != "" {
			related = append(related, videoId)
		}
	}
	return related
}
