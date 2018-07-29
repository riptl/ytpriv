package apijson

import (
	"github.com/valyala/fastjson"
	"github.com/terorie/yt-mango/data"
	"errors"
)

var missingData = errors.New("missing data")
var unexpectedType = errors.New("unexpected type")

func ParseVideo(v *data.Video, root *fastjson.Value) error {
	rootArray := root.GetArray()
	if rootArray == nil { return unexpectedType }

	var videoDetails *fastjson.Value
	for _, sub := range rootArray {
		videoDetails = sub.Get("page", "playerResponse", "videoDetails")
		if videoDetails != nil { break }
	}

	keywords := videoDetails.GetArray("keywords")
	if keywords == nil { return missingData }
	for _, keywordValue := range keywords {
		keywordBytes, _ := keywordValue.StringBytes()
		if keywordBytes == nil { continue }

		keyword := string(keywordBytes)
		v.Tags = append(v.Tags, keyword)
	}

	titleBytes := videoDetails.GetStringBytes("title")
	if titleBytes != nil { v.Title = string(titleBytes) }

	return nil
}
