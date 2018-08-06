package apis

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/api"
)

func TestClassicVideos(t *testing.T) {
	testVideos(t, &ClassicAPI)
}

func TestJsonVideos(t *testing.T) {
	testVideos(t, &JsonAPI)
}

func testVideos(t *testing.T, a *api.API) {
	testVideo1(t, a)
	testVideo2(t, a)
}

// Standard video test
func testVideo1(t *testing.T, a *api.API) {
	req := a.GrabVideo("uOXLKPCs54c")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	var v data.Video
	v.ID = "uOXLKPCs54c"
	recm, err := a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }
	assert.NotZero(t, len(recm), "No recommendations")

	assert.Equal(t, "https://www.youtube.com/watch?v=uOXLKPCs54c", v.URL)
	assert.Equal(t, 2013, v.UploadDate.Year())
	assert.Equal(t, time.October, v.UploadDate.Month())
	assert.Equal(t, 8, v.UploadDate.Day())
	if !(v.Duration == 10 || v.Duration == 11) {
		assert.Failf(t, "wrong duration", "expected: 10/11, actual: %d", v.Duration)
	}
	assert.Equal(t, "UCsLiV4WJfkTEHH0b9PmRklw", v.UploaderID)
	assert.Equal(t, "https://www.youtube.com/channel/UCsLiV4WJfkTEHH0b9PmRklw", v.UploaderURL)
	// TODO Formats
	assert.Equal(t, "Webdriver Torso", v.Uploader)
	assert.Equal(t, "tmpFmHnGe", v.Title)
	// TODO Description
	// TODO Thumbnail
	assert.Equal(t, "Creative Commons Attribution license (reuse allowed)", v.License)
	assert.Equal(t, "People & Blogs", v.Genre)
	// TODO Tags
	// TODO Subtitles
	assert.True(t, v.FamilyFriendly, "family friendly")
	assert.Equal(t, data.VisibilityPublic, v.Visibility)
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")
}

// Deleted video test
func testVideo2(t *testing.T, a *api.API) {
	req := a.GrabVideo("chGl0_nFyqg")

	res, err := net.Client.Do(req)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	v := data.Video{ID: "chGl0_nFyqg"}
	_, err = a.ParseVideo(&v, res)
	if err == nil {
		assert.FailNow(t, "no error on unavailable video")
	} else if err != api.VideoUnavailable {
		assert.FailNow(t, err.Error(), "wrong error thrown")
	}
}

// Age-restricted video test
func testVideo3(t *testing.T, a *api.API) {
	// i don't even know what this video is
	req := a.GrabVideo("r_tfGI7KBKA")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	v := data.Video{ID: "chGl0_nFyqg"}
	_, err = a.ParseVideo(&v, res)
	if err != nil { assert.FailNow(t, err.Error()) }

	assert.False(t, v.FamilyFriendly, "family friendly")
}
