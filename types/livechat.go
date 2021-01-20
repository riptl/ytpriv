package types

import "encoding/json"

// LivechatMessage is a message in YouTube live chats.
type LivechatMessage struct {
	ID        string          `json:"id"`
	Message   json.RawMessage `json:"message"`
	AuthorID  string          `json:"author_id"`
	Author    string          `json:"author"`
	Timestamp int64           `json:"timestamp"`
}

