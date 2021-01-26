package yt

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RequestPlaylistStart fetches the first page of a playlist.
func (c *Client) RequestPlaylistStart(id string) PlaylistStartRequest {
	const playlistURL = "https://www.youtube.com/playlist?pbj=1&list="
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(playlistURL + url.QueryEscape(id))
	setHeaders(&req.Header)
	return PlaylistStartRequest{c, req}
}

type PlaylistStartRequest struct {
	*Client
	*fasthttp.Request
}

func (r PlaylistStartRequest) Do() (*types.Playlist, error) {
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
	// TODO Check for unavailable videos
	for _, wrapper := range contents {
		renderer := wrapper.Get("playlistVideoRenderer")
		channel := renderer.Get("shortBylineText", "runs", "0")
		firstThumbnail := string(renderer.GetStringBytes("thumbnail", "thumbnails", "0", "url"))
		unavailable := firstThumbnail == "https://i.ytimg.com/img/no_thumbnail.jpg"
		video := types.VideoItem{
			ID:          string(renderer.GetStringBytes("videoId")),
			Title:       string(renderer.GetStringBytes("title", "runs", "0", "text")),
			ChannelID:   string(channel.GetStringBytes("navigationEndpoint", "browseEndpoint", "browseId")),
			ChannelName: string(channel.GetStringBytes("text")),
			Unavailable: unavailable,
		}
		if video.ID != "" {
			l.Page.Videos = append(l.Page.Videos, video)
		}
		if l.Page.Continuation == "" {
			l.Page.Continuation = string(wrapper.GetStringBytes("continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token"))
		}
	}
	return l, nil
}

// RequestPlaylistPage fetches a page of a playlist.
func (c *Client) RequestPlaylistPage(id string) PlaylistPageRequest {
	const uri = "https://www.youtube.com/youtubei/v1/browse?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.SetContentType("application/json")
	setHeaders(&req.Header)
	req.SetRequestURI(uri)
	var params = ytiRequest{
		Continuation: id,
		Context:      &defaultYTIContext,
	}
	reqBody, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	req.SetBody(reqBody)
	return PlaylistPageRequest{c, req}
}

type PlaylistPageRequest struct {
	*Client
	*fasthttp.Request
}

func (r PlaylistPageRequest) Do() (*types.PlaylistPage, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, err
	}
	return ParsePlaylistPage(res)
}

func ParsePlaylistPage(res *fasthttp.Response) (*types.PlaylistPage, error) {
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
	var p fastjson.Parser
	root, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}
	items := root.GetArray("onResponseReceivedActions", "0", "appendContinuationItemsAction", "continuationItems")
	var videos []types.VideoItem
	var continuation string
	for _, item := range items {
		renderer := item.Get("playlistVideoRenderer")
		videoID := string(renderer.GetStringBytes("videoId"))
		firstThumbnail := string(renderer.GetStringBytes("thumbnail", "thumbnails", "0", "url"))
		unavailable := firstThumbnail == "https://i.ytimg.com/img/no_thumbnail.jpg"
		title := string(renderer.GetStringBytes("title", "runs", "0", "text"))
		channelID := string(renderer.GetStringBytes("shortBylineText", "runs", "0", "navigationEndpoint", "browseEndpoint", "browseId"))
		channelName := string(renderer.GetStringBytes("shortBylineText", "runs", "0", "text"))
		if videoID != "" {
			videos = append(videos, types.VideoItem{
				ID:          videoID,
				Title:       title,
				ChannelID:   channelID,
				ChannelName: channelName,
				Unavailable: unavailable,
			})
		}
		if continuation == "" {
			continuation = string(item.GetStringBytes("continuationItemRenderer", "continuationEndpoint", "continuationCommand", "token"))
		}
	}
	return &types.PlaylistPage{
		Continuation: continuation,
		Videos:       videos,
	}, nil
}
