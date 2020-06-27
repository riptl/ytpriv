package api

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/terorie/ytwrk/data"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

func ParsePlaylist(l *data.Playlist, res *fasthttp.Response) error {
	// TODO Dedupe
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

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(res.Body())
	if err != nil {
		return err
	}
	rootArray := root.GetArray()
	if rootArray == nil {
		return unexpectedType
	}

	// Get interesting objects
	var response *fastjson.Value
	for _, sub := range rootArray {
		if r := sub.Get("response"); r != nil {
			response = r
			break
		}
	}
	if response == nil {
		return errors.New("no response")
	}

	sidebar := response.Get("sidebar", "playlistSidebarRenderer",
		"items", "0", "playlistSidebarPrimaryInfoRenderer")
	l.Views, _ = strconv.ParseInt(string(sidebar.GetStringBytes("stats", "1", "simpleText")), 10, 64)
	l.Title = string(sidebar.GetStringBytes("title", "runs", "0", "text"))

	playlist := response.Get("contents", "twoColumnBrowseResultsRenderer",
		"tabs", "0", "tabRenderer", "content", "sectionListRenderer",
		"contents", "0", "itemSectionRenderer", "contents", "0", "playlistVideoListRenderer")
	if playlist == nil {
		return errors.New("playlist not found")
	}
	contents := playlist.GetArray("contents")

	for _, wrapper := range contents {
		videoRenderer := wrapper.Get("playlistVideoRenderer")
		channel := videoRenderer.Get("shortBylineText", "runs", "0")
		video := data.PlaylistVideo{
			ID:          string(videoRenderer.GetStringBytes("videoId")),
			Title:       string(videoRenderer.GetStringBytes("title", "simpleText")),
			ChannelID:   string(channel.GetStringBytes("browseEndpoint", "browseId")),
			ChannelName: string(channel.GetStringBytes("text")),
		}
		l.Videos = append(l.Videos, video)
	}

	return nil
}
