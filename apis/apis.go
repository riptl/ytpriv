package apis

import (
	a "github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apiclassic"
	"github.com/terorie/yt-mango/apijson"
)

// TODO Fallback option
var Main *a.API = nil

// TODO: Remove when everything is implemented
var TempAPI = a.API{
	GrabVideo: apiclassic.GrabVideo,
	ParseVideo: apiclassic.ParseVideo,

	GrabChannel: apiclassic.GrabChannel,
	ParseChannel: apiclassic.ParseChannel,

	GrabChannelPage: apijson.GrabChannelPage,
	ParseChannelVideoURLs: apijson.ParseChannelVideoURLs,
}

var ClassicAPI = a.API{
	GrabVideo: apiclassic.GrabVideo,
	ParseVideo: apiclassic.ParseVideo,

	GrabChannel: apiclassic.GrabChannel,
	ParseChannel: apiclassic.ParseChannel,
}

var JsonAPI = a.API{
	GrabVideo: apijson.GrabVideo,
	ParseVideo: apijson.ParseVideo,

	GrabChannelPage: apijson.GrabChannelPage,
	ParseChannelVideoURLs: apijson.ParseChannelVideoURLs,
}

