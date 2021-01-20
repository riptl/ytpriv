package yt

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/terorie/ytwrk/types"
)

func TestClient_Video(t *testing.T) {
	client := NewClient()
	req := client.RequestVideo("uOXLKPCs54c")
	require.NotNil(t, req.Request)

	v, err := req.Do()
	require.NoError(t, err)
	require.NotNil(t, v)
	assert.NotZero(t, len(v.RelatedVideos), "No related videos")

	uploadDate := time.Unix(v.Uploaded, 0)
	assert.Equal(t, 2013, uploadDate.Year())
	assert.Equal(t, time.October, uploadDate.Month())
	assert.Equal(t, 8, uploadDate.Day())
	if !(v.Duration == 10 || v.Duration == 11) {
		assert.Failf(t, "wrong duration", "expected: 10/11, actual: %d", v.Duration)
	}
	assert.Equal(t, "UCsLiV4WJfkTEHH0b9PmRklw", v.UploaderID)
	assert.Equal(t, "Webdriver Torso", v.Uploader)
	assert.Equal(t, "tmpFmHnGe", v.Title)
	// TODO Thumbnail
	assert.Equal(t, "Creative Commons Attribution license (reuse allowed)", v.License)
	assert.Equal(t, "People & Blogs", v.Genre)
	// TODO Subtitles
	assert.True(t, v.FamilyFriendly, "family friendly")
	assert.Equal(t, types.VisibilityPublic, v.Visibility, "visibility")
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")

	// Check formats
	formats := []string{"18", "134", "243", "133", "242", "160", "278", "140", "251"}
	assert.Equal(t, formats, v.Formats)
}

func TestClient_Video_Deleted(t *testing.T) {
	v, err := NewClient().RequestVideo("chGl0_nFyqg").Do()
	assert.EqualError(t, err, "video unavailable")
	assert.Nil(t, v)
}

// Age-restricted video test
func TestClient_Video_Restricted(t *testing.T) {
	v, err := NewClient().RequestVideo("6kLq3WMV1nU").Do()
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, "Dedication To My Ex (Miss That) (Lyric Video)", v.Title)
	assert.Equal(t, "LloydVEVO", v.Uploader)
	assert.Equal(t, "UCYvy_rZWF3udXeDHy3PBdtw", v.UploaderID)
	uploadDate := time.Unix(v.Uploaded, 0)
	assert.Equal(t, 2011, uploadDate.Year())
	assert.Equal(t, time.June, uploadDate.Month())
	assert.Equal(t, 29, uploadDate.Day())
	assert.False(t, v.FamilyFriendly, "family friendly")

	// Parse tags
	tagList := []string{
		"Lloyd", "new", "video", "Dedication", "to",
		"my", "Ex", "Lil", "Wayne", "Andre", "3000",
		"Cupid", "King", "of", "Hearts",
	}
	tagMap := make(map[string]int)
	// Register every tag as 2
	for _, tag := range tagList {
		tagMap[tag] = 2
	}
	// Increment every tag
	for _, tag := range v.Tags {
		tagMap[tag]++
	}
	// Every tag should be 3
	for _, tag := range tagList {
		assert.NotEqual(t, 1, "Unknown tag", tag)
		assert.NotEqual(t, 2, "Missing tag", tag)
	}
}

// Description test
func TestClient_Video_Description(t *testing.T) {
	v, err := NewClient().RequestVideo("kj9mFK62c6E").Do()
	require.NoError(t, err)
	require.NotNil(t, v)

	const descTest4start = `never buy think pad hyper;

Test for project:
https://example.org/ gud link
some unicode: ¥≈ç√∫~
some attacks: `

	const descTest4stop = ` \n

*a* _b_ -c-`

	// YouTube as of 2018-08-08 has a bug where
	// some tokens are not escaped properly.
	// If a user includes the tokens "&lt;" or "&gt;"
	// in the description literally,
	// they are not escaped (e.g. to "&amp;lt;")
	// and show up as "<" or ">" in the page source
	// on "/watch".
	assert.True(t, strings.HasPrefix(v.Description, descTest4start))
	println(v.Description)
	assert.True(t, strings.HasSuffix(v.Description, descTest4stop))
}

// Unlisted video test
func TestClient_Video_Unlisted(t *testing.T) {
	v, err := NewClient().RequestVideo("RCz83F_OGsI").Do()
	require.NoError(t, err)
	require.NotNil(t, v)

	assert.Equal(t, "Houdini Smoke Sim Test", v.Title)
	assert.Equal(t, "Henry Medhurst", v.Uploader)
	assert.Equal(t, "UCbYbnbrBh4mULa3fjfwJVgw", v.UploaderID)
	uploadDate := time.Unix(v.Uploaded, 0)
	assert.Equal(t, 2014, uploadDate.Year())
	assert.Equal(t, time.April, uploadDate.Month())
	assert.Equal(t, 8, uploadDate.Day())
	assert.Equal(t, "Film & Animation", v.Genre)
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")
	assert.EqualValues(t, types.VisibilityUnlisted, v.Visibility, "visibility")
}

