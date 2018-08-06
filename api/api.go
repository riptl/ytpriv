package api

import (
	"github.com/terorie/yt-mango/data"
	"net/http"
)

type API struct {
	// Build a request to grab the video page
	GrabVideo func(videoID string) *http.Request
	// Parse a response with a video page into a video struct
	// and return all related video IDs or an error
	ParseVideo func(*data.Video, *http.Response) ([]string, error)

	GrabVideoSubtitleList func(videoID string) *http.Request
	ParseVideoSubtitleList func(*data.Video, *http.Response) error

	GrabChannel func(channelID string) *http.Request
	ParseChannel func(*data.Channel, *http.Response) error

	GrabChannelPage func(channelID string, page uint) *http.Request
	ParseChannelVideoURLs func(*http.Response) ([]string, error)
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
