package yt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLivechatReplay(t *testing.T) {
	c := NewClient()
	// Request video containing livechat replay.
	vid, err := c.RequestVideo("-_YBydPNp70").Do()
	require.NoError(t, err)
	assert.NotEmpty(t, vid.LiveChatReplayContinuation)
	assert.NotEmpty(t, vid.TopChatReplayContinuation)
	// Get first page.
	msgs1, cont1, err := c.RequestLivechatReplay(vid.LiveChatReplayContinuation).Do()
	require.NoError(t, err)
	require.NotEmpty(t, cont1)
	require.NotEmpty(t, msgs1)
	t.Log("Page 1")
	for _, msg := range msgs1 {
		t.Log("Msg", string(msg.Message))
	}
	// Get second page.
	msgs2, cont2, err := c.RequestLivechatReplay(cont1).Do()
	require.NoError(t, err)
	assert.NotEmpty(t, cont2)
	require.NotEmpty(t, msgs2)
	assert.NotEqual(t, msgs1[0].ID, msgs2[0].ID)
	t.Log("Page 2")
	for _, msg := range msgs2 {
		t.Log("Msg", string(msg.Message))
	}
}
