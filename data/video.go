package data

type Video struct {
	// Static data
	ID         string   `json:"id"`
	Uploaded   int64    `json:"uploaded,omitempty"`
	Duration   uint64   `json:"duration"`
	UploaderID string   `json:"uploader_id"`
	Formats    []string `json:"formats,omitempty"`

	// Content metadata
	Uploader       string      `json:"uploader"` // The channel name can change
	Title          string      `json:"title"`
	Description    string      `json:"description"`
	License        string      `json:"license,omitempty"`
	Genre          string      `json:"genre"`
	Tags           []string    `json:"tags,omitempty"`
	Captions       []Caption   `json:"captions,omitempty"`
	FamilyFriendly bool        `json:"family_friendly"`
	Livestream     *Livestream `json:"livestream,omitempty"`

	// Privacy settings
	Visibility       VisibilitySetting `json:"visibility"`
	NoComments       bool              `json:"no_comments,omitempty"`
	NoRatings        bool              `json:"no_ratings,omitempty"`
	NoEmbed          bool              `json:"no_embed,omitempty"`
	ProductPlacement bool              `json:"product_placement,omitempty"`
	WatchStatistics  bool              `json:"watch_statistics,omitempty"`

	// Dynamic stats
	Views         uint64         `json:"views"`
	Likes         uint64         `json:"likes"`
	Dislikes      uint64         `json:"dislikes"`
	RelatedVideos []RelatedVideo `json:"related"`

	// Internal tokens
	Internal interface{} `json:"-"`
}

type RelatedVideo struct {
	ID         string `json:"id"`
	UploaderID string `json:"uploader_id"`
}

type Livestream struct {
	LowLatency   bool `json:"low_latency"`
	DvrEnabled   bool `json:"dvr_enabled"`
	OwnerViewing bool `json:"owner_viewing"`
	LiveContent  bool `json:"is_live_content"`
}

type Caption struct {
	VssID        string `json:"vss_id"`
	Name         string `json:"name"`
	Code         string `json:"code"`
	Translatable bool   `json:"translatable"`
}

type VisibilitySetting uint8

const (
	VisibilityPublic = VisibilitySetting(iota)
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
