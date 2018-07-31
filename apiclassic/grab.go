package apiclassic

import (
	"net/http"
	"errors"
	"encoding/xml"
	"github.com/terorie/yt-mango/net"
	"fmt"
)

const videoURL = "https://www.youtube.com/watch?has_verified=1&bpctr=6969696969&v="
const subtitleURL = "https://video.google.com/timedtext?type=list&v="
const channelURL = "https://www.youtube.com/channel/%s/about"

func GrabVideo(videoID string) *http.Request {
	req, err := http.NewRequest("GET", videoURL + videoID, nil)
	if err != nil { panic(err) }
	setHeaders(&req.Header)

	return req
}

func GrabSubtitleList(videoID string) (tracks *XMLSubTrackList, err error) {
	req, err := http.NewRequest("GET", subtitleURL + videoID, nil)
	if err != nil { return }
	setHeaders(&req.Header)

	res, err := net.Client.Do(req)
	if err != nil { return }
	if res.StatusCode != 200 { return nil, errors.New("HTTP failure") }

	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)

	tracks = new(XMLSubTrackList)
	err = decoder.Decode(tracks)
	return
}

func GrabChannel(channelID string) *http.Request {
	req, err := http.NewRequest("GET", fmt.Sprintf(channelURL, channelID), nil)
	if err != nil { panic(err) }
	setHeaders(&req.Header)

	return req
}

func setHeaders(h *http.Header) {
	h.Add("Host", "www.youtube.com")
	h.Add("User-Agent", "yt-mango/0.1")
}
