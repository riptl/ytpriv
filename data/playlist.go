package data

type Playlist struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Views  int64           `json:"views"`
	Videos []PlaylistVideo `json:"videos"`
}

type PlaylistVideo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	ChannelID   string `json:"name"`
	ChannelName string `json:"channel_name"`
}
