package apijson

import (
	"github.com/terorie/yt-mango/data"
	"errors"
)

func GetVideo(v *data.Video) (err error) {
	jsn, err := GrabVideo(v)
	if err != nil { return }
	err = ParseVideo(v, jsn)
	if err != nil { return }
	return
}

func GetChannel(c *data.Channel) error {
	return errors.New("not implemented")
}

func GetChannelVideoURLs(channelID string, page uint) (urls []string, err error) {
	jsn, err := GrabChannelPage(channelID, page)
	if err != nil { return }
	urls, err = ParseChannelPageLinks(jsn)
	return
}
