package apijson

import (
	"github.com/valyala/fasthttp"
)

const videoURL = "https://www.youtube.com/watch?pbj=1&v="
const channelURL = "https://www.youtube.com/browse_ajax?ctoken="

func GrabVideo(videoID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(videoURL + videoID)
	setHeaders(&req.Header)

	return req
}

func GrabChannelPage(channelID string, page uint) *fasthttp.Request {
	// Generate page URL
	token := GenChannelPageToken(channelID, uint64(page))
	url := channelURL + token

	// Prepare request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	setHeaders(&req.Header)

	return req
}

func setHeaders(h *fasthttp.RequestHeader) {
	h.Add("Host", "www.youtube.com")
	h.Add("X-YouTube-Client-Name", "1")
	h.Add("X-YouTube-Client-Version", "2.20170707")
}
