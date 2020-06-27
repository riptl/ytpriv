package data

import "encoding/json"

type Comment struct {
	// Static
	ID             string `json:"id"`
	VideoID        string `json:"video_id"`
	AuthorID       string `json:"author_id"`
	ByChannelOwner bool   `json:"by_channel_owner,omitempty"`
	ParentID       string `json:"parent_id,omitempty"`

	// Static target: Accuracy-improving changes
	CreatedText   string `json:"created_text"`
	CreatedBefore int64  `json:"created_before"`
	CreatedAfter  int64  `json:"created_after"`

	// Malleable: Author-triggered changes
	Author  string          `json:"author"`
	Content json.RawMessage `json:"content"`
	Edited  bool            `json:"edited,omitempty"`

	// Dynamic
	CrawledAt  int64  `json:"crawled_at"`
	LikeCount  uint64 `json:"likes"`
	ReplyCount uint64 `json:"replies,omitempty"`

	// Application-specific (not exported)
	Internal interface{} `json:"-"`
}
