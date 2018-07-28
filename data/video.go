package data

import "time"

type Video struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Description string `json:"description"`
	Uploader string `json:"uploader"`
	UploaderID string `json:"uploader_id"`
	UploaderURL string `json:"uploader_url"`
	UploadDate time.Time `json:"upload_date"`
	Thumbnail string `json:"thumbnail"`
	URL string `json:"url"`
	License string `json:"license,omitempty"`
	Genre string `json:"genre"`
	Tags []string `json:"tags,omitempty"`
	Subtitles []string `json:"subtitles,omitempty"`
	Duration uint64 `json:"duration"`
	FamilyFriendly bool `json:"family_friendly"`
	Views uint64 `json:"views"`
	Likes uint64 `json:"likes"`
	Dislikes uint64 `json:"dislikes"`
	Formats []Format `json:"formats,omitempty"`
}

type Subtitle struct {
	URL string
	Extension string
}
