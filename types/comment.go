package types

import "encoding/json"

// A Comment on a YouTube video.
type Comment struct {
	// Static
	ID             string
	VideoID        string
	AuthorID       string
	ByChannelOwner bool   `json:",omitempty"`
	ParentID       string `json:",omitempty"`

	// Static target: Accuracy-improving changes
	CreatedText   string
	CreatedBefore int64
	CreatedAfter  int64

	// Malleable: Author-triggered changes
	Author  string
	Content json.RawMessage
	Edited  bool `json:",omitempty"`

	// Dynamic
	CrawledAt  int64
	LikeCount  uint64
	ReplyCount uint64 `json:",omitempty"`

	// Application-specific (not exported)
	Internal interface{} `json:"-"`
}

type CommentPage struct {
	Comments         []Comment
	CommentParseErrs []error
	MoreComments     bool
	Continuation     *CommentContinuation
	TopComments      *CommentContinuation
	NewComments      *CommentContinuation
}

type CommentContinuation struct {
	VideoID  string
	ParentID string
	Cookie   string
	Token    string
	XSRF     string
}
