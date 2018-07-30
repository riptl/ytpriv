package apijson

import (
	"github.com/valyala/fastjson"
	"github.com/terorie/yt-mango/data"
	"errors"
	"io/ioutil"
	"net/http"
)

var missingData = errors.New("missing data")
var unexpectedType = errors.New("unexpected type")

func ParseVideo(v *data.Video, res *http.Response) error {
	defer res.Body.Close()

	// Download response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return err }

	// Parse JSON
	var p fastjson.Parser
	root, err := p.ParseBytes(body)
	if err != nil { return err }

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
