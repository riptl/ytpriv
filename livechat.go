package yt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

// RequestLivechat fetches the continuation of a live chat.
func (c *Client) RequestLivechat(continuation string) LivechatRequest {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	setHeaders(&req.Header)
	req.SetRequestURI("https://www.youtube.com/youtubei/v1/live_chat/get_live_chat?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8")
	var params = ytiRequest{
		Context:      &defaultYTIContext,
		Continuation: continuation,
	}
	reqBody, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	req.SetBody(reqBody)
	return LivechatRequest{c, req}
}

type LivechatRequest struct {
	*Client
	*fasthttp.Request
}

func (r LivechatRequest) Do() ([]*types.LivechatMessage, LivechatContinuation, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, LivechatContinuation{}, err
	}
	return ParseLivechat(res)
}

func ParseLivechat(res *fasthttp.Response) (msgs []*types.LivechatMessage, cont LivechatContinuation, err error) {
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
	contents := root.Get("continuationContents")
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
		// Live chat direct action.
		msg := parseLivechatMessage(action)
		if msg != nil {
			parsed = append(parsed, msg)
		}
		// Wrapped replay action.
		for _, innerAction := range action.GetArray("replayChatItemAction", "actions") {
			msg := parseLivechatMessage(innerAction)
			if msg != nil {
				parsed = append(parsed, msg)
			}
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
		paidMsgRuns := paidMsg.Get("message", "runs")
		msg := &types.LivechatMessage{
			ID:         id,
			AuthorID:   string(paidMsg.GetStringBytes("authorExternalChannelId")),
			Author:     string(paidMsg.GetStringBytes("authorName", "simpleText")),
			Timestamp:  timestamp,
			SuperChat:  true,
			PaidAmount: string(paidMsg.GetStringBytes("purchaseAmountText", "simpleText")),
		}
		if paidMsgRuns != nil {
			msg.Message = paidMsgRuns.MarshalTo(nil)
		}
		return msg
	}
	return nil
}

type LivechatContinuation struct {
	Timeout      int
	Continuation string
}
