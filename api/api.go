package api

import (
	"github.com/terorie/yt-mango/data"
	"net/http"
	"github.com/terorie/yt-mango/apijson"
	"github.com/terorie/yt-mango/apiclassic"
)

type API struct {
	GrabVideo func(videoID string) *http.Request
	ParseVideo func(*data.Video, *http.Response) error

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
