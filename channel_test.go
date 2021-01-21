package yt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terorie/ytwrk/types"
)

func TestClient_ChannelOverview(t *testing.T) {
	client := NewClient()
	req := client.RequestChannelOverview("UCV6mNrW8CrmWtcxWfQXy11g")
	require.NotNil(t, req.Request)
	overview, err := req.Do()
	require.NoError(t, err)
	assert.Equal(t, &types.ChannelOverview{
		ChannelID: "UCV6mNrW8CrmWtcxWfQXy11g",
		Title:     "DarkViperAU",
		Links: types.ChannelHeaderLinks{
			Twitch:  "https://www.twitch.tv/DarkViperAU",
			Twitter: "https://www.twitter.com/DarkViperAU",
			Patreon: "http://www.patreon.com/darkviperau",
			Discord: "https://discord.gg/DarkViperAU",
			TikTok:  "https://vm.tiktok.com/9Hbyea/",
		},
		Verified:  true,
		Sponsored: true,
	}, overview)
}

func TestClient_ChannelVideos(t *testing.T) {
	client := NewClient()
	// Start request
	req1 := client.RequestChannelVideosStart("UCV6mNrW8CrmWtcxWfQXy11g")
	require.NotNil(t, req1.Request)
	page1, err := req1.Do()
	require.NoError(t, err)
	require.NotNil(t, page1)
	assert.NotEmpty(t, page1.Continuation, "no continuation")
	assert.Len(t, page1.Videos, 30, "video count")
	for _, video := range page1.Videos {
		assert.NotEmpty(t, video.ID, "empty video ID")
		assert.NotEmpty(t, video.Title, "empty video title")
		assert.NotEmpty(t, video.ChannelID, "empty channel ID")
		assert.NotEmpty(t, video.ChannelName, "empty channel name")
	}
	// Continuation request
	req2 := client.RequestChannelVideosPage(page1.Continuation)
	require.NotNil(t, req2.Request)
	page2, err := req2.Do()
	require.NoError(t, err)
	require.NotNil(t, page2)
	assert.NotEmpty(t, page2.Continuation, "no continuation")
	assert.Len(t, page2.Videos, 30, "video count")
	for _, video := range page2.Videos {
		assert.NotEmpty(t, video.ID, "empty video ID")
		assert.NotEmpty(t, video.Title, "empty video title")
		assert.NotEmpty(t, video.ChannelID, "empty channel ID")
		assert.NotEmpty(t, video.ChannelName, "empty channel name")
	}
}
