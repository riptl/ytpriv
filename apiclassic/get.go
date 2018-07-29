package apiclassic

import (
	"github.com/terorie/yt-mango/data"
	"errors"
)

func GetVideo(v *data.Video) error {
	if len(v.ID) == 0 { return errors.New("no video ID") }

	// Download the doc tree
	doc, err := grab(v)
	if err != nil { return err }

	// Parse it
	p := parseInfo{v, doc}
	err = p.parse()
	if err != nil { return err }

	return nil
}

func GetChannel(c *data.Channel) error {
	return errors.New("not implemented")
}

func GetChannelVideoURLs(channelID string, page uint) ([]string, error) {
	return nil, errors.New("not implemented")
}
