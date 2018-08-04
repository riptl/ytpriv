package api

import (
	"github.com/terorie/yt-mango/data"
	"net/http"
	"github.com/terorie/yt-mango/apijson"
	"github.com/terorie/yt-mango/apiclassic"
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

// TODO Fallback option
var Main *API = nil

// TODO: Remove when everything is implemented
var TempAPI = API{
	GrabVideo: apiclassic.GrabVideo,
	ParseVideo: apiclassic.ParseVideo,

	GrabChannel: apiclassic.GrabChannel,
	ParseChannel: apiclassic.ParseChannel,

	GrabChannelPage: apijson.GrabChannelPage,
	ParseChannelVideoURLs: apijson.ParseChannelVideoURLs,
}

var ClassicAPI = API{
	GrabVideo: apiclassic.GrabVideo,
	ParseVideo: apiclassic.ParseVideo,

	GrabChannel: apiclassic.GrabChannel,
	ParseChannel: apiclassic.ParseChannel,
}

var JsonAPI = API{
	GrabVideo: apijson.GrabVideo,
	ParseVideo: apijson.ParseVideo,

	GrabChannelPage: apijson.GrabChannelPage,
	ParseChannelVideoURLs: apijson.ParseChannelVideoURLs,
}
