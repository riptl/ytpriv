package types

type Playlist struct {
	ID    string
	Title string
	Views int64
	Page  PlaylistPage
}

type PlaylistPage struct {
	Continuation string
	Videos       []VideoItem
}

// VideoItem is a video in a list reference.
type VideoItem struct {
	ID          string
	Title       string
	ChannelID   string `json:",omitempty"`
	ChannelName string `json:",omitempty"`
	Unavailable bool   `json:",omitempty"`
}
