package yt

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"

	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RequestLivechatStart fetches the beginning of a live chat.
func (c *Client) RequestLivechatStart(videoID string) LivechatStartRequest {
	const livechatURL = "https://www.youtube.com/live_chat?pbj=1&v="
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(livechatURL + url.QueryEscape(videoID))
	setHeaders(&req.Header)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15")
	return LivechatStartRequest{c, req}
}

type LivechatStartRequest struct {
	*Client
	*fasthttp.Request
}

func (r LivechatStartRequest) Do() ([]*types.LivechatMessage, LivechatContinuation, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, LivechatContinuation{}, err
	}
	return ParseLivechatStart(res)
}

func ParseLivechatStart(res *fasthttp.Response) (msgs []*types.LivechatMessage, cont LivechatContinuation, err error) {
	if res.StatusCode() != fasthttp.StatusOK {
		err = fmt.Errorf("response status %d", res.StatusCode())
		return
	}
	contentType := res.Header.ContentType()
	if !bytes.HasPrefix(contentType, []byte("application/json")) {
		err = ErrRateLimit
		return
	}
	var parser fastjson.Parser
	var root *fastjson.Value
	root, err = parser.ParseBytes(res.Body())
	if err != nil {
		return
	}
	main := root.Get("1", "response")
	contents := main.Get("contents")
	liveChatRenderer := contents.Get("liveChatRenderer")
	continuationObj := liveChatRenderer.Get("continuations", "0", "timedContinuationData")
	cont = LivechatContinuation{
		Timeout:      continuationObj.GetInt("timeoutMs"),
		Continuation: string(continuationObj.GetStringBytes("continuation")),
	}
	chatMessages := liveChatRenderer.GetArray("actions")
	msgs = parseLiveChatMessages(chatMessages)
	return
}

// RequestLivechatContinuation fetches the continuation of a live chat.
func (c *Client) RequestLivechatContinuation(continuation string) LivechatContinuationRequest {
	const livechatPageURL = "https://www.youtube.com/live_chat/get_live_chat?pbj=1&hidden=false&continuation="
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod("GET")
	req.SetRequestURI(livechatPageURL + continuation)
	setHeaders(&req.Header)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Safari/605.1.15")
	return LivechatContinuationRequest{c, req}
}

type LivechatContinuationRequest struct {
	*Client
	*fasthttp.Request
}

func (r LivechatContinuationRequest) Do() ([]*types.LivechatMessage, LivechatContinuation, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, LivechatContinuation{}, err
	}
	return ParseLivechatPage(res)
}

func ParseLivechatPage(res *fasthttp.Response) (msgs []*types.LivechatMessage, cont LivechatContinuation, err error) {
	if res.StatusCode() != fasthttp.StatusOK {
		err = fmt.Errorf("response status %d", res.StatusCode())
		return
	}
	contentType := res.Header.ContentType()
	if !bytes.HasPrefix(contentType, []byte("application/json")) {
		err = ErrRateLimit
		return
	}
	var parser fastjson.Parser
	var root *fastjson.Value
	root, err = parser.ParseBytes(res.Body())
	if err != nil {
		return
	}
	main := root.Get("response")
	contents := main.Get("continuationContents")
	liveChatRenderer := contents.Get("liveChatContinuation")
	continuationObj := liveChatRenderer.Get("continuations", "0", "timedContinuationData")
	cont = LivechatContinuation{
		Timeout:      continuationObj.GetInt("timeoutMs"),
		Continuation: string(continuationObj.GetStringBytes("continuation")),
	}
	chatMessages := liveChatRenderer.GetArray("actions")
	msgs = parseLiveChatMessages(chatMessages)
	return
}

func parseLiveChatMessages(actions []*fastjson.Value) (parsed []*types.LivechatMessage) {
	for _, action := range actions {
		msg := parseLivechatMessage(action)
		if msg != nil {
			parsed = append(parsed, msg)
		}
	}
	return
}

func parseLivechatMessage(action *fastjson.Value) *types.LivechatMessage {
	// A normal chat message.
	chatMsg := action.Get("addChatItemAction", "item", "liveChatTextMessageRenderer")
	if chatMsg != nil {
		timestampStr := string(chatMsg.GetStringBytes("timestampUsec"))
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		id, _ := url.QueryUnescape(string(chatMsg.GetStringBytes("id")))
		return &types.LivechatMessage{
			ID:        id,
			Message:   chatMsg.Get("message", "runs").MarshalTo(nil),
			AuthorID:  string(chatMsg.GetStringBytes("authorExternalChannelId")),
			Author:    string(chatMsg.GetStringBytes("authorName", "simpleText")),
			Timestamp: timestamp,
		}
	}
	// A paid/super-chat message.
	paidMsg := action.Get("addChatItemAction", "item", "liveChatPaidMessageRenderer")
	if paidMsg != nil {
		timestampStr := string(paidMsg.GetStringBytes("timestampUsec"))
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		id, _ := url.QueryUnescape(string(paidMsg.GetStringBytes("id")))
		return &types.LivechatMessage{
			ID:         id,
			Message:    paidMsg.Get("message", "runs").MarshalTo(nil),
			AuthorID:   string(paidMsg.GetStringBytes("authorExternalChannelId")),
			Author:     string(paidMsg.GetStringBytes("authorName", "simpleText")),
			Timestamp:  timestamp,
			SuperChat:  true,
			PaidAmount: string(paidMsg.GetStringBytes("purchaseAmountText", "simpleText")),
		}
	}
	return nil
}

type LivechatContinuation struct {
	Timeout      int
	Continuation string
}
