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
	ChannelID   string
	ChannelName string
	Unavailable bool `json:",omitempty"`
}
