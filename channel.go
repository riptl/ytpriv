package yt

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

func (c *Client) RequestChannelOverview(channelID string) ChannelOverviewRequest {
	const uri = "https://www.youtube.com/youtubei/v1/browse?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	setHeaders(&req.Header)
	req.SetRequestURI(uri)
	var params = ytiRequest{
		BrowseID: channelID,
		Context:  &defaultYTIContext,
	}
	reqBody, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	req.SetBody(reqBody)
	return ChannelOverviewRequest{c, req}
}

type ChannelOverviewRequest struct {
	*Client
	*fasthttp.Request
}

func (r ChannelOverviewRequest) Do() (*types.ChannelOverview, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, err
	}
	return ParseChannelOverview(res)
}

func ParseChannelOverview(res *fasthttp.Response) (*types.ChannelOverview, error) {
	if res.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}
	var p fastjson.Parser
	rootObj, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}
	overview := new(types.ChannelOverview)
	c4Header := rootObj.Get("header", "c4TabbedHeaderRenderer")
	overview.ChannelID = string(c4Header.GetStringBytes("channelId"))
	overview.Title = string(c4Header.GetStringBytes("title"))
	for _, link := range c4Header.GetArray("headerLinks", "channelHeaderLinksRenderer", "primaryLinks") {
		parseLink(overview, link)
	}
	for _, link := range c4Header.GetArray("headerLinks", "channelHeaderLinksRenderer", "secondaryLinks") {
		parseLink(overview, link)
	}
	for _, badge := range c4Header.GetArray("badges") {
		if string(badge.GetStringBytes("metadataBadgeRenderer", "style")) == "BADGE_STYLE_TYPE_VERIFIED" {
			overview.Verified = true
		}
	}
	if c4Header.Get("sponsorButton", "buttonRenderer") != nil {
		overview.Sponsored = true
	}
	return overview, nil
}

func parseLink(overview *types.ChannelOverview, v *fastjson.Value) {
	kind := string(v.GetStringBytes("title", "simpleText"))
	endpoint := string(v.GetStringBytes("navigationEndpoint", "urlEndpoint", "url"))
	endpointURI, err := url.Parse(endpoint)
	if err != nil {
		return
	}
	link := endpointURI.Query().Get("q")
	if link == "" {
		return
	}
	switch kind {
	case "Twitch":
		overview.Links.Twitch = link
	case "Twitter":
		overview.Links.Twitter = link
	case "Patreon":
		overview.Links.Patreon = link
	case "Reddit":
		overview.Links.Reddit = link
	case "Discord":
		overview.Links.Discord = link
	case "TikTok":
		overview.Links.TikTok = link
	}
}

func (c *Client) RequestChannelVideosStart(channelID string) ChannelVideosStartRequest {
	const uri = "https://www.youtube.com/youtubei/v1/browse?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	setHeaders(&req.Header)
	req.SetRequestURI(uri)
	var params = ytiRequest{
		BrowseID: channelID,
		Context:  &defaultYTIContext,
		Params:   "EgZ2aWRlb3M%3D",
	}
	reqBody, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	req.SetBody(reqBody)
	return ChannelVideosStartRequest{c, req}
}

type ChannelVideosStartRequest struct {
	*Client
	*fasthttp.Request
}

func (r ChannelVideosStartRequest) GetRequest() *fasthttp.Request {
	return r.Request
}

func (r ChannelVideosStartRequest) Do() (*types.ChannelVideosPage, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, err
	}
	return ParseChannelVideosStart(res)
}

func ParseChannelVideosStart(res *fasthttp.Response) (*types.ChannelVideosPage, error) {
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}
	var p fastjson.Parser
	rootObj, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}
	c4Header := rootObj.Get("header", "c4TabbedHeaderRenderer")
	channelID := string(c4Header.GetStringBytes("channelId"))
	channelName := string(c4Header.GetStringBytes("title"))
	tabs := rootObj.GetArray("contents", "twoColumnBrowseResultsRenderer", "tabs")
	var gridRenderer *fastjson.Value
	for _, tab := range tabs {
		gridRenderer = tab.Get("tabRenderer", "content", "sectionListRenderer", "contents", "0", "itemSectionRenderer", "contents", "0", "gridRenderer")
		if gridRenderer != nil {
			break
		}
	}
	if gridRenderer == nil {
		return nil, fmt.Errorf("video list not found")
	}
	page := new(types.ChannelVideosPage)
	page.Continuation = string(gridRenderer.GetStringBytes("continuations", "0", "nextContinuationData", "continuation"))
	for _, item := range gridRenderer.GetArray("items") {
		renderer := item.Get("gridVideoRenderer")
		page.Videos = append(page.Videos, types.VideoItem{
			ID:          string(renderer.GetStringBytes("videoId")),
			Title:       string(renderer.GetStringBytes("title", "runs", "0", "text")),
			ChannelID:   channelID,
			ChannelName: channelName,
		})
	}
	return page, nil
}

// RequestChannelVideosPage requests a page of videos on a channel, given a continuation.
func (c *Client) RequestChannelVideosPage(cont string) ChannelVideosPageRequest {
	const baseURI = "https://www.youtube.com/browse_ajax?"
	uri := baseURI + url.Values{
		"ctoken":       {cont},
		"continuation": {cont},
	}.Encode()
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	setHeaders(&req.Header)
	return ChannelVideosPageRequest{c, req}
}

type ChannelVideosPageRequest struct {
	*Client
	*fasthttp.Request
}

func (r ChannelVideosPageRequest) GetRequest() *fasthttp.Request {
	return r.Request
}

func (r ChannelVideosPageRequest) Do() (*types.ChannelVideosPage, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.HTTP.Do(r.Request, res); err != nil {
		return nil, err
	}
	return ParseChannelVideosPage(res)
}

func ParseChannelVideosPage(res *fasthttp.Response) (*types.ChannelVideosPage, error) {
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}
	var p fastjson.Parser
	rootObj, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}
	root, err := rootObj.Array()
	if err != nil {
		return nil, err
	}
	var container *fastjson.Value
	for _, item := range root {
		if item.Exists("response") {
			container = item
			break
		}
	}
	if container == nil {
		return nil, MissingData
	}
	errorExists := container.Exists("response", "responseContext", "errors", "error")
	if errorExists {
		return nil, ServerError
	}
	grid := container.Get("response", "continuationContents", "gridContinuation")
	continuation := string(grid.GetStringBytes("continuations", "0", "nextContinuationData", "continuation"))
	channelMeta := container.Get("response", "metadata", "channelMetadataRenderer")
	channelID := string(channelMeta.GetStringBytes("externalId"))
	channelName := string(channelMeta.GetStringBytes("title"))
	var videos []types.VideoItem
	for _, item := range grid.GetArray("items") {
		videos = append(videos, types.VideoItem{
			ID:          string(item.GetStringBytes("gridVideoRenderer", "videoId")),
			Title:       string(item.GetStringBytes("gridVideoRenderer", "title", "runs", "0", "text")),
			ChannelID:   channelID,
			ChannelName: channelName,
		})
	}
	return &types.ChannelVideosPage{
		Continuation: continuation,
		Videos:       videos,
	}, nil
}
