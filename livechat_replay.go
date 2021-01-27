package yt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/terorie/ytpriv/types"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

func (c *Client) RequestLivechatReplay(continuation string) LivechatReplayRequest {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodPost)
	req.Header.SetContentType("application/json")
	setHeaders(&req.Header)
	// TODO Where does this key come from?
	req.SetRequestURI("https://www.youtube.com/youtubei/v1/live_chat/get_live_chat_replay?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8")
	var params = ytiRequest{
		Context:      &defaultYTIContext,
		Continuation: continuation,
	}
	reqBody, err := json.Marshal(&params)
	if err != nil {
		panic(err)
	}
	req.SetBody(reqBody)
	return LivechatReplayRequest{c, req}
}

type LivechatReplayRequest struct {
	*Client
	*fasthttp.Request
}

func (r LivechatReplayRequest) Do() ([]*types.LivechatMessage, string, error) {
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := r.Client.HTTP.Do(r.Request, res); err != nil {
		return nil, "", err
	}
	return ParseLivechatReplay(res)
}

func ParseLivechatReplay(res *fasthttp.Response) (msgs []*types.LivechatMessage, cont string, err error) {
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
	continuations := root.GetArray("continuationContents", "liveChatContinuation", "continuations")
	for _, contWrapper := range continuations {
		cont = string(contWrapper.GetStringBytes("liveChatReplayContinuationData", "continuation"))
		if cont != "" {
			break
		}
	}
	chatMessages := root.GetArray("continuationContents", "liveChatContinuation", "actions")
	msgs = parseLiveChatMessages(chatMessages)
	return
}
