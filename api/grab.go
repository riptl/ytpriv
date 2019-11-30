package api

import (
	"encoding/xml"
	"fmt"
	"net/url"

	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

const videoURL = "https://www.youtube.com/watch?pbj=1&v="
const channelURL = "https://www.youtube.com/channel/%s/about?pbj=1"
const channelPageURL = "https://www.youtube.com/browse_ajax?ctoken="
const commentURL = "https://www.youtube.com/comment_service_ajax?action_get_comments=1&pbj=1&ctoken=%[1]s&continuation=%[1]s"
const subtitleURL = "https://video.google.com/timedtext?type=list&v="
const livechatURL = "https://www.youtube.com/live_chat?pbj=1&v="
const livechatPageURL = "https://www.youtube.com/live_chat/get_live_chat?pbj=1&hidden=false&continuation="
const playlistURL = "https://www.youtube.com/playlist?pbj=1&list="

func GrabVideo(videoID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(videoURL + url.QueryEscape(videoID))
	setHeaders(&req.Header)

	return req
}

func GrabChannel(channelID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(fmt.Sprintf(channelURL, url.PathEscape(channelID)))
	setHeaders(&req.Header)

	return req
}

func GrabCommentPage(continuation *CommentContinuation) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("POST")
	req.Header.Set("Cookie", continuation.Cookie)
	req.SetRequestURI(fmt.Sprintf(commentURL, continuation.Token))
	req.SetBodyString("session_token=" + continuation.XSRF)
	setHeaders(&req.Header)
	return req
}

func GrabChannelPage(channelID string, page uint) *fasthttp.Request {
	// Generate page URL
	token := GenChannelPageToken(channelID, uint64(page))
	uri := channelPageURL + url.QueryEscape(token)

	// Prepare request
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(uri)
	setHeaders(&req.Header)

	return req
}

func GrabVideoSubtitleList(videoID string) (tracks *XMLSubTrackList, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(subtitleURL + videoID)
	setXMLHeaders(&req.Header)

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

func GrabLiveChatStart(videoID string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(livechatURL + url.QueryEscape(videoID))
	setHeaders(&req.Header)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15")
	return req
}

func GrabLiveChatContinuation(continuation string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(livechatPageURL + continuation)
	setHeaders(&req.Header)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15")
	return req
}

func GrabPlaylist(id string) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(playlistURL + url.QueryEscape(id))
	setHeaders(&req.Header)
	return req
}

func setHeaders(h *fasthttp.RequestHeader) {
	h.Add("Accept-Language", "en-US")
	h.Add("Host", "www.youtube.com")
	h.Add("X-YouTube-Client-Name", "1")
	h.Add("X-YouTube-Client-Version", "2.20170707")
}

func setXMLHeaders(h *fasthttp.RequestHeader) {
	h.Add("Accept-Language", "en-US")
	h.Add("Host", "www.youtube.com")
}

type XMLSubTrackList struct {
	Tracks []struct {
		LangCode string `xml:"lang_code,attr"`
		Lang     string `xml:"lang_translated,attr"`
	} `xml:"track"`
}
