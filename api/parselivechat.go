package api

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"

	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
)

func ParseLiveChatStart(out *[]data.LiveChatMessage, res *fasthttp.Response) (data.LiveChatContinuation, error) {
	var cont data.LiveChatContinuation

	if res.StatusCode() != fasthttp.StatusOK {
		return cont, fmt.Errorf("response status %d", res.StatusCode())
	}
	contentType := res.Header.ContentType()
	if !bytes.HasPrefix(contentType, []byte("application/json")) {
		return cont, ErrRateLimit
	}

	var parser fastjson.Parser
	root, err := parser.ParseBytes(res.Body())
	if err != nil {
		return cont, err
	}

	main := root.Get("1", "response")
	contents := main.Get("contents")
	liveChatRenderer := contents.Get("liveChatRenderer")
	continuationObj := liveChatRenderer.Get("continuations", "0", "timedContinuationData")
	cont = data.LiveChatContinuation{
		Timeout:      continuationObj.GetInt("timeoutMs"),
		Continuation: string(continuationObj.GetStringBytes("continuation")),
	}
	chatMessages := liveChatRenderer.GetArray("actions")
	*out, err = parseLiveChatMessages(chatMessages)

	return cont, err
}

func ParseLiveChatPage(out *[]data.LiveChatMessage, res *fasthttp.Response) (data.LiveChatContinuation, error) {
	var cont data.LiveChatContinuation

	if res.StatusCode() != fasthttp.StatusOK {
		return cont, fmt.Errorf("response status %d", res.StatusCode())
	}
	contentType := res.Header.ContentType()
	if !bytes.HasPrefix(contentType, []byte("application/json")) {
		return cont, ErrRateLimit
	}

	var parser fastjson.Parser
	root, err := parser.ParseBytes(res.Body())
	if err != nil {
		return cont, err
	}

	main := root.Get("response")
	contents := main.Get("continuationContents")
	liveChatRenderer := contents.Get("liveChatContinuation")
	continuationObj := liveChatRenderer.Get("continuations", "0", "timedContinuationData")
	cont = data.LiveChatContinuation{
		Timeout:      continuationObj.GetInt("timeoutMs"),
		Continuation: string(continuationObj.GetStringBytes("continuation")),
	}
	chatMessages := liveChatRenderer.GetArray("actions")
	*out, err = parseLiveChatMessages(chatMessages)

	return cont, err
}

func parseLiveChatMessages(actions []*fastjson.Value) (parsed []data.LiveChatMessage, err error) {
	for _, chatMessage := range actions {
		messageRenderer := chatMessage.Get("addChatItemAction", "item", "liveChatTextMessageRenderer")
		if messageRenderer == nil {
			continue
		}
		timestampStr := string(messageRenderer.GetStringBytes("timestampUsec"))
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		id, _ := url.QueryUnescape(string(messageRenderer.GetStringBytes("id")))
		parsed = append(parsed, data.LiveChatMessage{
			ID:        id,
			Message:   messageRenderer.Get("message", "runs").MarshalTo(nil),
			AuthorID:  string(messageRenderer.GetStringBytes("authorExternalChannelId")),
			Author:    string(messageRenderer.GetStringBytes("authorName", "simpleText")),
			Timestamp: timestamp,
		})
	}
	return
}
