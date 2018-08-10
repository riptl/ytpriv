package data

import (
	"time"
)

type Video struct {
	// Static data
	ID string `json:"id" bson:"id"`
	URL string `json:"url" bson:"url"`
	UploadDate time.Time `json:"upload_date" bson:"upload_date"`
	Duration uint64 `json:"duration" bson:"duration"`
	UploaderID string `json:"uploader_id" bson:"uploader_id"`
	UploaderURL string `json:"uploader_url" bson:"uploader_url"`
	Formats []string `json:"formats,omitempty" bson:"formats,omitempty"`

	// Content metadata
	Uploader string `json:"uploader" bson:"uploader"` // The channel name can change
	Title string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Thumbnail string `json:"thumbnail" bson:"thumbnail"`
	License string `json:"license,omitempty" bson:"license,omitempty"`
	Genre string `json:"genre" bson:"genre"`
	Tags []string `json:"tags,omitempty" bson:"tags"`
	Subtitles []string `json:"subtitles,omitempty" bson:"subtitles,omitempty"`
	FamilyFriendly bool `json:"family_friendly" bson:"family_friendly"`

	// Privacy settings
	Visibility VisibilitySetting `json:"visibility" bson:"visibility"`
	NoComments bool `json:"no_comments" bson:"no_comments"`
	NoRatings bool `json:"no_ratings" bson:"no_ratings"`
	NoEmbed bool `json:"no_embed" bson:"no_embed"`
	ProductPlacement bool `json:"product_placement" bson:"product_placement"`
	WatchStatistics bool `json:"watch_statistics" bson:"watch_statistics"`

	// Dynamic stats
	Views uint64 `json:"views" bson:"views"`
	Likes uint64 `json:"likes" bson:"likes"`
	Dislikes uint64 `json:"dislikes" bson:"dislikes"`
}

type Subtitle struct {
	URL string
	Extension string
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
