package data

import "encoding/json"

type LiveChatMessage struct {
	ID        string          `json:"id"`
	Message   json.RawMessage `json:"message"`
	AuthorID  string          `json:"author_id"`
	Author    string          `json:"author"`
	Timestamp int64           `json:"timestamp"`
}

type LiveChatContinuation struct {
	Timeout int
	Continuation string
}
