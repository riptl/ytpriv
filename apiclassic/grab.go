package apiclassic

import (
	"encoding/xml"
	"fmt"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

const videoURL = "https://www.youtube.com/watch?has_verified=1&bpctr=6969696969&v="
const subtitleURL = "https://video.google.com/timedtext?type=list&v="
const channelURL = "https://www.youtube.com/channel/%s/about"

func GrabVideo(videoID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(videoURL + videoID)
	setHeaders(&req.Header)

	return req
}

func GrabSubtitleList(videoID string) (tracks *XMLSubTrackList, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(subtitleURL + videoID)
	setHeaders(&req.Header)

	res := fasthttp.AcquireResponse()

	err = net.Client.Do(req, res)
	if err != nil { return }

	if res.StatusCode() != 200 {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}

	tracks = new(XMLSubTrackList)
	err = xml.Unmarshal(res.Body(), tracks)

	return
}

func GrabChannel(channelID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fmt.Sprintf(channelURL, channelID))
	setHeaders(&req.Header)

	return req
}

func setHeaders(h *fasthttp.RequestHeader) {
	h.Add("Accept-Language", "en-US")
	h.Add("Host", "www.youtube.com")
}
