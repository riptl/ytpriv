package apiclassic

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const likeBtnSelector = ".like-button-renderer-like-button-unclicked"
const dislikeBtnSelector = ".like-button-renderer-dislike-button-unclicked"
const userInfoSelector = "div .yt-user-info"
const channelNameSelector = ".yt-uix-sessionlink"
const recommendSelector = ".related-list-item"

var playerConfigErr = errors.New("failed to parse player config")

func ParseVideo(v *data.Video, res *fasthttp.Response) (err error) {
	if res.StatusCode() != 200 {
		return fmt.Errorf("HTTP status %d", res.StatusCode())
	}

	buf := res.Body()
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(buf))
	if err != nil { return }

	p := parseVideoInfo{v: v, doc: doc}
	return p.parse()
}

type parseVideoInfo struct {
	v *data.Video
	doc *goquery.Document
	restricted bool
}

func (p *parseVideoInfo) parse() error {
	available, err := p.isAvailable()
	if err != nil { return err }
	if !available { return api.VideoUnavailable }

	if err := p.parseLikeDislike();
		err != nil { return err }
	if err := p.parseUploader();
		err != nil { return err }
	if err := p.parseDescription();
		err != nil { return err }
	if err := p.parsePlayerConfig();
		err != nil { return err }
	if err := p.parseMetas();
		err != nil { return err }
	if err := p.parseRecommends();
		err != nil { return err }
	p.parseLicense()
	return nil
}

func (p *parseVideoInfo) isAvailable() (bool, error) {
	// Get the player-unavailable warning and check if it's hidden
	unavTag := p.doc.Find("#player-unavailable")
	if len(unavTag.Nodes) != 1 {
		return false, errors.New("cannot check if player is available")
	}

	// Warning is hidden: Must be available
	if unavTag.HasClass("hid") { return true, nil }

	// Video is either restricted or deleted from here …
	p.restricted = true

	// Find the <meta name="title"… tag.
	// If it's empty, then the video is unavailable
	var unav bool
	p.doc.Find("meta").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if s.AttrOr("name", "") == "title" {
			unav = s.AttrOr("content", "") == ""
			return false
		}
		return true
	})

	return !unav, nil
}

func (p *parseVideoInfo) parseLikeDislike() error {
	likeText := p.doc.Find(likeBtnSelector).First().Text()
	dislikeText := p.doc.Find(dislikeBtnSelector).First().Text()

	if len(likeText) == 0 || len(dislikeText) == 0 {
		return errors.New("failed to parse like buttons")
	}

	var err error
	p.v.Likes, err = api.ExtractNumber(likeText)
	if err != nil { return err }
	p.v.Dislikes, err = api.ExtractNumber(dislikeText)
	if err != nil { return err }

	return nil
}

func (p *parseVideoInfo) parseUploader() error {
	userInfo := p.doc.Find(userInfoSelector)
	userLinkNode := userInfo.Find(".yt-uix-sessionlink")

	// get link
	userLink, _ := userLinkNode.Attr("href")
	if userLink == "" { return errors.New("couldn't find channel link") }
	p.v.UploaderURL = "https://www.youtube.com" + userLink

	// get name
	channelName := userInfo.Find(channelNameSelector).Text()
	if channelName == "" { return errors.New("could not find channel name") }
	p.v.Uploader = channelName
	return nil
}

func (p *parseVideoInfo) parseMetas() (err error) {
	enumMetas(p.doc.Selection, func(tag metaTag)bool {
		content := tag.content
		switch tag.typ {
		case metaProperty:
			switch tag.name {
			case "og:title":
				p.v.Title = content
			case "og:video:tag":
				p.v.Tags = append(p.v.Tags, content)
			case "og:url":
				p.v.URL = content
			case "og:image":
				p.v.Thumbnail = content
			}
		case metaName:
			switch tag.name {
			}
		case metaItemProp:
			switch tag.name {
			case "datePublished":
				if val, err := time.Parse("2006-01-02", content);
					err == nil { p.v.UploadDate = val }
			case "genre":
				p.v.Genre = content
			case "channelId":
				p.v.UploaderID = content
			case "duration":
				if val, err := api.ParseDuration(content); err == nil {
					p.v.Duration = val
				} else {
					return false
				}
			case "isFamilyFriendly":
				if val, err := strconv.ParseBool(content);
					err == nil { p.v.FamilyFriendly = val }
			case "interactionCount":
				if val, err := strconv.ParseUint(content, 10, 64);
					err == nil { p.v.Views = val }
			case "unlisted":
				if val, err := strconv.ParseBool(content);
					err == nil && val { p.v.Visibility = data.VisibilityUnlisted }
			}
		}
		return true
	})

	return err
}

func (p *parseVideoInfo) parsePlayerConfig() error {
	// Player config is unavailable on restricted vids
	if p.restricted { return nil }

	var json string

	p.doc.Find("script").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		script := s.Text()
		startMatch := regexp.MustCompile("var ytplayer = ytplayer \\|\\| {};\\s*ytplayer\\.config = {")
		endMatch := regexp.MustCompile("};\\s*ytplayer.load = function\\(")

		startIndices := startMatch.FindStringIndex(script)
		if startIndices == nil { return true }
		endIndices := endMatch.FindStringIndex(script)
		if endIndices == nil { return true }

		// minus one to preserve braces
		startIndex, endIndex := startIndices[1] - 1, endIndices[0] + 1
		if startIndex > endIndex { return true }

		json = script[startIndex:endIndex]

		// Stop searching, json found
		return false
	})
	// No json found
	if json == "" { return playerConfigErr }

	// Try decoding json
	var parser fastjson.Parser
	config, err := parser.Parse(json)
	if err != nil { return err }

	// Extract data
	args := config.Get("args")
	if args == nil { return playerConfigErr }

	// Get fmt_list string
	fmts := string(args.GetStringBytes("fmt_list"))
	if fmts == "" { return playerConfigErr }

	// Split and decode it
	fmtList, err := api.ParseFormatList(fmts)
	if err != nil { return err }
	p.v.Formats = fmtList

	return nil
}

func (p *parseVideoInfo) parseRecommends() error {
	s := p.doc.Find(recommendSelector).Find(".content-wrapper").Find("a")
	s.Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists { return }
		if !strings.HasPrefix(href, "/watch?v=") { return }
		id := href[len("/watch?v="):]
		p.v.Related = append(p.v.Related, id)
	})
	return nil
}

func (p *parseVideoInfo) parseLicense() {
	p.doc.Find(".watch-meta-item").EachWithBreak(func(i int, s *goquery.Selection) bool {
		title := strings.Trim(s.Find("h4").Text(), "\n ")
		if title == "License" {
			p.v.License = s.Find("a").Text()
			return false
		}
		return true
	})
}
