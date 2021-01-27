package types

// A Video on YouTube.
type Video struct {
	// Static data
	ID         string
	Uploaded   int64 `json:",omitempty"`
	Duration   uint64
	UploaderID string
	Formats    []string `json:",omitempty"`

	// Content metadata
	Uploader       string
	Title          string
	Description    string
	License        string `json:",omitempty"`
	Genre          string
	Tags           []string  `json:",omitempty"`
	Captions       []Caption `json:",omitempty"`
	FamilyFriendly bool
	Livestream     *Livestream `json:",omitempty"`

	// Continuations
	TopChatContinuation        string `json:",omitempty"`
	LiveChatContinuation       string `json:",omitempty"`
	TopChatReplayContinuation  string `json:",omitempty"`
	LiveChatReplayContinuation string `json:",omitempty"`

	// Privacy settings
	Visibility       VisibilitySetting
	NoComments       bool `json:",omitempty"`
	NoRatings        bool `json:",omitempty"`
	NoEmbed          bool `json:",omitempty"`
	ProductPlacement bool `json:",omitempty"`
	WatchStatistics  bool `json:",omitempty"`

	// Dynamic stats
	Views         uint64
	Likes         uint64
	Dislikes      uint64
	RelatedVideos []RelatedVideo

	// Internal tokens
	Internal interface{} `json:"-"`
}

type RelatedVideo struct {
	ID         string
	UploaderID string
}

type Livestream struct {
	LowLatency   bool
	DvrEnabled   bool
	OwnerViewing bool
	LiveContent  bool
}

type Caption struct {
	VssID        string
	Name         string
	Code         string
	Translatable bool
}

type VisibilitySetting uint8

const (
	_ = VisibilitySetting(iota)
	VisibilityPublic
	VisibilityUnlisted
	VisibilityPrivate
)

func (t VisibilitySetting) MarshalJSON() ([]byte, error) {
	var str string
	switch t {
	case VisibilityPublic:
		str = "\"public\""
	case VisibilityUnlisted:
		str = "\"unlisted\""
	case VisibilityPrivate:
		str = "\"private\""
	}
	return []byte(str), nil
}
