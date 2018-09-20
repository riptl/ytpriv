package apis

import (
	"github.com/stretchr/testify/assert"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"strings"
	"testing"
	"time"
)

func TestClassicVideo(t *testing.T) { testVideo(t, &ClassicAPI) }
func TestClassicVideoDeleted(t *testing.T) { testVideoDeleted(t, &ClassicAPI) }
func TestClassicVideoRestricted(t *testing.T) { testVideoRestricted(t, &ClassicAPI) }
func TestClassicVideoDescription(t *testing.T) { testVideoDescription(t, &ClassicAPI) }
func TestClassicVideoUnlisted(t *testing.T) { testVideoUnlisted(t, &ClassicAPI) }

func TestJsonVideo(t *testing.T) { testVideo(t, &JsonAPI) }
func TestJsonVideoDeleted(t *testing.T) { testVideoDeleted(t, &JsonAPI) }
func TestJsonVideoRestricted(t *testing.T) { testVideoRestricted(t, &JsonAPI) }
func TestJsonVideoDescription(t *testing.T) { testVideoDescription(t, &JsonAPI) }
func TestJsonVideoUnlisted(t *testing.T) { testVideoUnlisted(t, &JsonAPI) }

// Standard video test
func testVideo(t *testing.T, a *api.API) {
	req := a.GrabVideo("uOXLKPCs54c")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	var v data.Video
	v.ID = "uOXLKPCs54c"
	err = a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }
	assert.NotZero(t, len(v.Related), "No recommendations")

	assert.Equal(t, "https://www.youtube.com/watch?v=uOXLKPCs54c", v.URL)
	assert.Equal(t, 2013, v.UploadDate.Year())
	assert.Equal(t, time.October, v.UploadDate.Month())
	assert.Equal(t, 8, v.UploadDate.Day())
	if !(v.Duration == 10 || v.Duration == 11) {
		assert.Failf(t, "wrong duration", "expected: 10/11, actual: %d", v.Duration)
	}
	assert.Equal(t, "UCsLiV4WJfkTEHH0b9PmRklw", v.UploaderID)
	assert.Equal(t, "https://www.youtube.com/channel/UCsLiV4WJfkTEHH0b9PmRklw", v.UploaderURL)
	assert.Equal(t, "Webdriver Torso", v.Uploader)
	assert.Equal(t, "tmpFmHnGe", v.Title)
	// TODO Thumbnail
	assert.Equal(t, "Creative Commons Attribution license (reuse allowed)", v.License)
	assert.Equal(t, "People & Blogs", v.Genre)
	// TODO Subtitles
	assert.True(t, v.FamilyFriendly, "family friendly")
	assert.Equal(t, data.VisibilityPublic, v.Visibility)
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")

	// Check formats
	formats := []string{"43", "18", "36", "17"}
	assert.Equal(t, len(formats), len(v.Formats), "format count")
	for i, fmt := range v.Formats {
		assert.Equal(t, formats[i], fmt, "format ID", i)
	}
}

// Deleted video test
func testVideoDeleted(t *testing.T, a *api.API) {
	req := a.GrabVideo("chGl0_nFyqg")

	res, err := net.Client.Do(req)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	v := data.Video{ID: "chGl0_nFyqg"}
	err = a.ParseVideo(&v, res)
	if err == nil {
		assert.FailNow(t, "no error on unavailable video")
	} else if err != api.VideoUnavailable {
		assert.FailNow(t, err.Error(), "wrong error thrown")
	}
}

// Age-restricted video test
func testVideoRestricted(t *testing.T, a *api.API) {
	req := a.GrabVideo("6kLq3WMV1nU")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	v := data.Video{ID: "6kLq3WMV1nU"}
	// Age-restricted vids don't have recommendations
	err = a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }

	assert.Equal(t, "Dedication To My Ex (Miss That) (Lyric Video)", v.Title)
	assert.Equal(t, "LloydVEVO", v.Uploader)
	assert.Equal(t, "UCYvy_rZWF3udXeDHy3PBdtw", v.UploaderID)
	assert.Equal(t, 2011, v.UploadDate.Year())
	assert.Equal(t, time.June, v.UploadDate.Month())
	assert.Equal(t, 29, v.UploadDate.Day())
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
func testVideoDescription(t *testing.T, a *api.API) {
	req := a.GrabVideo("kj9mFK62c6E")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	v := data.Video{ID: "kj9mFK62c6E"}
	err = a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }

	const descTest4start =
`never buy think pad hyper;

Test for project:
https://example.org/ gud link
some unicode: ¥≈ç√∫~
some attacks: `

	const descTest4stop =
` \n

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
func testVideoUnlisted(t *testing.T, a *api.API) {
	req := a.GrabVideo("RD5otQyBFqc")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	v := data.Video{ID: "RD5otQyBFqc"}
	err = a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }

	assert.Equal(t, "How Northern Lights Are Created", v.Title)
	assert.Equal(t, "Love Nature", v.Uploader)
	assert.Equal(t, "UCRZPkuHwaoKwTP3CYPdVldg", v.UploaderID)
	assert.Equal(t, 2013, v.UploadDate.Year())
	assert.Equal(t, time.February, v.UploadDate.Month())
	assert.Equal(t, 11, v.UploadDate.Day())
	assert.Equal(t, "Science & Technology", v.Genre)
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")
	assert.EqualValues(t, data.VisibilityUnlisted, v.Visibility, "visibility")
}
