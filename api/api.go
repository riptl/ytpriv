package api

import (
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/apiclassic"
	"github.com/terorie/yt-mango/apijson"
)

type API struct {
	GetVideo func(*data.Video) error
	GetVideoSubtitleList func(*data.Video) error
	GetChannel func(*data.Channel) error
	GetChannelVideoURLs func(channelID string, page uint) ([]string, error)
}

// TODO Fallback option
var DefaultAPI *API = nil

var ClassicAPI = API{
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
}
