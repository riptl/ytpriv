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
)

var matchThumbUrl = regexp.MustCompile("^.+/hqdefault\\.jpg")

var missingData = errors.New("missing data")
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
		}
		if playerArgs == nil {
			playerArgs = sub.Get("player", "args")
		}
	}
	if playerResponse == nil { return errors.New("no video details") }
	if playerArgs == nil { return errors.New("no player args") }

	// Playable in embed?
	playableInEmbedValue := playerResponse.Get("playabilityStatus", "playableInEmbed")
	if playableInEmbedValue.Exists() {
		playableInEmbed, _ := playableInEmbedValue.Bool()
		v.NoEmbed = !playableInEmbed
	}

	if err := parseVideoDetails(v, playerResponse.Get("videoDetails"));
		err != nil { return err }

	watchNextContents := pageResponse.GetArray(
		"contents", "twoColumnWatchNextResults", "results", "results", "contents")
	if err := parseVideoInfo(v, watchNextContents);
		err != nil { return err }

	// Get URL
	v.URL = string(playerArgs.GetStringBytes("loaderUrl"))

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
	if err != nil { v.Duration = length64 }

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

	// Get upload date
	dateText := string(secondary.GetStringBytes("dateText", "simpleText"))
	dateText = strings.TrimPrefix(dateText, "Published on ")
	date, err := time.Parse("Jan _2, 2006", dateText)
	if err == nil { v.UploadDate = data.SimpleTime(date) }

	return nil
}

