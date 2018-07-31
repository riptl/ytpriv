package apijson

import (
	"net/http"
)

const videoURL = "https://www.youtube.com/watch?pbj=1&v="
const channelURL = "https://www.youtube.com/browse_ajax?ctoken="

func GrabVideo(videoID string) *http.Request {
	req, err := http.NewRequest("GET", videoURL + videoID, nil)
	if err != nil { panic(err) }
	setHeaders(&req.Header)

	return req
}

func GrabChannelPage(channelID string, page uint) *http.Request {
	// Generate page URL
	token := GenChannelPageToken(channelID, uint64(page))
	url := channelURL + token

	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { panic(err) }
	setHeaders(&req.Header)

	return req
}

func setHeaders(h *http.Header) {
	h.Add("Host", "www.youtube.com")
	h.Add("User-Agent", "yt-mango/0.1")
	h.Add("X-YouTube-Client-Name", "1")
	h.Add("X-YouTube-Client-Version", "2.20170707")
}
