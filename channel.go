package yt

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RequestChannel requests a channel by ID.
func (c *Client) RequestChannel(channelID string) *fasthttp.Request {
	const channelURL = "https://www.youtube.com/channel/%s/about?pbj=1"
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fmt.Sprintf(channelURL, url.PathEscape(channelID)))
	setHeaders(&req.Header)
	return req
}

type ChannelRequest struct {
	*Client
	*fasthttp.Request
}

// RequestChannelPage requests a page of videos on a channel.
func (c *Client) RequestChannelPage(channelID string, page uint) *fasthttp.Request {
	const channelPageURL = "https://www.youtube.com/browse_ajax?ctoken="
	token := GenChannelPageToken(channelID, uint64(page))
	uri := channelPageURL + url.QueryEscape(token)
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	setHeaders(&req.Header)
	return req
}

func ParseChannelVideoURLs(res *fasthttp.Response) ([]string, error) {
	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}

	// Parse JSON
	var p fastjson.Parser
	rootObj, err := p.ParseBytes(res.Body())
	if err != nil {
		return nil, err
	}

	// Root as array
	root, err := rootObj.Array()
	if err != nil {
		return nil, err
	}

	// Find response container
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

	// Get error obj
	errorExists := container.Exists(
		"response",
		"responseContext",
		"errors",
		"error",
	)
	if errorExists {
		return nil, ServerError
	}

	// Get items from grid
	itemsObj := container.Get(
		"response",
		"continuationContents",
		"gridContinuation",
		"items",
	)
	// End of data
	if itemsObj == nil {
		return nil, nil
	}

	// Items as array
	items, err := itemsObj.Array()
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0)

	// Enumerate
	for _, item := range items {
		// Find URL
		urlObj := item.Get(
			"gridVideoRenderer",
			"navigationEndpoint",
			"commandMetadata",
			"webCommandMetadata",
			"url",
		)
		if urlObj == nil {
			return nil, MissingData
		}

		// URL as string
		urlBytes, err := urlObj.StringBytes()
		if err != nil {
			return nil, err
		}
		urlStr := string(urlBytes)
		if strings.HasPrefix(urlStr, "/watch?v") {
			urls = append(urls, "https://www.youtube.com"+urlStr)
		}
	}
	return urls, nil
}
