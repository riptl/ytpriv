package yt

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/terorie/ytwrk/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RequestPlaylist fetches the first page of a playlist.
func (c *Client) RequestPlaylist(id string) PlaylistRequest {
	const playlistURL = "https://www.youtube.com/playlist?pbj=1&list="
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(playlistURL + url.QueryEscape(id))
	setHeaders(&req.Header)
	return PlaylistRequest{c, req}
}

type PlaylistRequest struct {
	*Client
	*fasthttp.Request
}

func (r PlaylistRequest) Do() (*types.Playlist, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, err
	}
	return ParsePlaylist(res)
}

func ParsePlaylist(res *fasthttp.Response) (*types.Playlist, error) {
	l := new(types.Playlist)
	// TODO Dedupe
	if res.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("response status %d", res.StatusCode())
	}
	contentType := res.Header.ContentType()
	switch {
	case bytes.HasPrefix(contentType, []byte("application/json")):
		break
	case bytes.HasPrefix(contentType, []byte("text/html")):
		return nil, ErrRateLimit
	}

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}
	rootArray := root.GetArray()
	if rootArray == nil {
		return nil, unexpectedType
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
		return nil, errors.New("no response")
	}

	sidebar := response.Get("sidebar", "playlistSidebarRenderer",
		"items", "0", "playlistSidebarPrimaryInfoRenderer")
	l.Views, _ = strconv.ParseInt(string(sidebar.GetStringBytes("stats", "1", "simpleText")), 10, 64)
	l.Title = string(sidebar.GetStringBytes("title", "runs", "0", "text"))

	playlist := response.Get("contents", "twoColumnBrowseResultsRenderer",
		"tabs", "0", "tabRenderer", "content", "sectionListRenderer",
		"contents", "0", "itemSectionRenderer", "contents", "0", "playlistVideoListRenderer")
	if playlist == nil {
		return nil, errors.New("playlist not found")
	}
	contents := playlist.GetArray("contents")

	for _, wrapper := range contents {
		videoRenderer := wrapper.Get("playlistVideoRenderer")
		channel := videoRenderer.Get("shortBylineText", "runs", "0")
		video := types.PlaylistVideo{
			ID:          string(videoRenderer.GetStringBytes("videoId")),
			Title:       string(videoRenderer.GetStringBytes("title", "simpleText")),
			ChannelID:   string(channel.GetStringBytes("browseEndpoint", "browseId")),
			ChannelName: string(channel.GetStringBytes("text")),
		}
		l.Videos = append(l.Videos, video)
	}

	return nil, nil
}
