package data

import "time"

type Video struct {
	// Static data
	ID string `json:"id"`
	URL string `json:"url"`
	UploadDate SimpleTime `json:"upload_date"`
	Duration uint64 `json:"duration"`
	UploaderID string `json:"uploader_id"`
	UploaderURL string `json:"uploader_url"`
	Formats []Format `json:"formats,omitempty"`

	// Content metadata
	Uploader string `json:"uploader"` // The channel name can change
	Title string `json:"title"`
	Description string `json:"description"`
	Thumbnail string `json:"thumbnail"`
	License string `json:"license,omitempty"`
	Genre string `json:"genre"`
	Tags []string `json:"tags,omitempty"`
	Subtitles []string `json:"subtitles,omitempty"`
	FamilyFriendly bool `json:"family_friendly"`

	// Privacy settings
	Visibility VisibilitySetting `json:"visibility"`
	NoComments bool `json:"no_comments"`
	NoRatings bool `json:"no_ratings"`
	NoEmbed bool `json:"no_embed"`
	ProductPlacement bool `json:"product_placement"`
	WatchStatistics bool `json:"watch_statistics"`

	// Dynamic stats
	Views uint64 `json:"views"`
	Likes uint64 `json:"likes"`
	Dislikes uint64 `json:"dislikes"`
}

type Subtitle struct {
	URL string
	Extension string
}

type SimpleTime time.Time

func (t SimpleTime) MarshalJSON() ([]byte, error) {
	return []byte(time.Time(t).Format("\"2006-01-02\"")), nil
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
		case VisibilityPublic: str =  "\"public\""
		case VisibilityUnlisted: str = "\"unlisted\""
		case VisibilityPrivate: str = "\"private\""
	}
	return []byte(str), nil
}
