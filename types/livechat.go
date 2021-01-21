package types

import "encoding/json"

// LivechatMessage is a message or super chat in YouTube live chats.
type LivechatMessage struct {
	ID        string
	Message   json.RawMessage
	AuthorID  string
	Author    string
	Timestamp int64

	// Super Chat specific
	SuperChat  bool
	PaidAmount string
}
