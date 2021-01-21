package yt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_PlaylistVideos(t *testing.T) {
	client := NewClient()
	// Start request
	req1 := client.RequestPlaylistStart("PL2satA_B-xnSAxmFXHgb1tsaVJ_Pfhrg2")
	require.NotNil(t, req1.Request)
	page1, err := req1.Do()
	require.NoError(t, err)
	require.NotNil(t, page1)
	require.NotEmpty(t, page1.Page.Continuation, "no continuation")
	assert.Len(t, page1.Page.Videos, 100, "video count")
	hasChannelID := false
	hasChannelName := false
	for _, video := range page1.Page.Videos {
		assert.NotEmpty(t, video.ID, "empty video ID")
		assert.NotEmpty(t, video.Title, "empty video title")
		hasChannelID = hasChannelID || (video.ChannelID != "")
		hasChannelName = hasChannelName || (video.ChannelName != "")
	}
	assert.True(t, hasChannelID, "no video had a channel ID")
	assert.True(t, hasChannelName, "no video had a channel name")
	// Next page
	req2 := client.RequestPlaylistPage(page1.Page.Continuation)
	require.NotNil(t, req2.Request)
	page2, err := req2.Do()
	require.NoError(t, err)
	require.NotNil(t, page2)
	assert.NotEmpty(t, page2.Continuation, "no continuation")
	assert.Len(t, page1.Page.Videos, 100, "video count")
	hasChannelID = false
	hasChannelName = false
	for _, video := range page2.Videos {
		assert.NotEmpty(t, video.ID, "empty video ID")
		assert.NotEmpty(t, video.Title, "empty video title")
		hasChannelID = hasChannelID || (video.ChannelID != "")
		hasChannelName = hasChannelName || (video.ChannelName != "")
	}
	assert.True(t, hasChannelID, "no video had a channel ID")
	assert.True(t, hasChannelName, "no video had a channel name")
}
