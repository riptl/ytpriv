package types

type ChannelOverview struct {
	ChannelID string
	Title     string
	Links     ChannelHeaderLinks
	Verified  bool
	Sponsored bool
}

type ChannelHeaderLinks struct {
	Twitch  string `json:",omitempty"`
	Twitter string `json:",omitempty"`
	Patreon string `json:",omitempty"`
	Reddit  string `json:",omitempty"`
	Discord string `json:",omitempty"`
	TikTok  string `json:",omitempty"`
}

type ChannelVideosPage struct {
	Continuation string
	Videos       []VideoItem
}
