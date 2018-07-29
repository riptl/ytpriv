package api

import (
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/classic"
	"github.com/terorie/yt-mango/apiclassic"
)

type API struct {
	GetVideo func(*data.Video) error
	GetChannel func(*data.Channel) error
	GetChannelVideoURLs func(channelID string, page uint) ([]string, error)
}

var ClassicAPI = API{
	GetVideo: apiclassic.GetVideo,
	GetChannel: apiclassic.GetChannel,
	GetChannelVideoURLs: apiclassic.GetChannelVideoURLs,
}

var JsonAPI struct {

}
