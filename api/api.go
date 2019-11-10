package api

import (
	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fasthttp"
)

type API struct {
	// Build a request to grab the video page
	GrabVideo func(videoID string) *fasthttp.Request
	// Parse a response with a video page into a video struct
	ParseVideo func(*data.Video, *fasthttp.Response) error

	GrabChannel func(channelID string) *fasthttp.Request
	ParseChannel func(*data.Channel, *fasthttp.Response) error

	GrabChannelPage func(channelID string, page uint) *fasthttp.Request
	ParseChannelVideoURLs func(*fasthttp.Response) ([]string, error)
}

type Err int
const (
	GenericError = Err(iota)
	VideoUnavailable
)

func (e Err) Error() string { switch e {
	case VideoUnavailable:
		return "video unavailable"
	default:
		return "unknown error"
}}
