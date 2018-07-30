package api

import (
	"github.com/terorie/yt-mango/data"
	"net/http"
	"github.com/terorie/yt-mango/apijson"
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
	GrabVideo: apijson.GrabVideo,
	ParseVideo: apijson.ParseVideo,

	GrabChannelPage: apijson.GrabChannelPage,
	ParseChannelVideoURLs: apijson.ParseChannelVideoURLs,
}

/*var ClassicAPI = API{
	GetVideo: apiclassic.GetVideo,
	GetVideoSubtitleList: apiclassic.GetVideoSubtitleList,
	GetChannel: apiclassic.GetChannel,
	GetChannelVideoURLs: apiclassic.GetChannelVideoURLs,
}

var JsonAPI = API{
	GetVideo: apijson.GetVideo,
	GetVideoSubtitleList: apiclassic.GetVideoSubtitleList,
	GetChannel: apijson.GetChannel,
	GetChannelVideoURLs: apijson.GetChannelVideoURLs,
}*/
