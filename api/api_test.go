package api

import (
	"testing"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestClassicVideo(t *testing.T) {
	testVideo(t, &ClassicAPI)
}

func TestJsonVideo(t *testing.T) {
	testVideo(t, &JsonAPI)
}

func testVideo(t *testing.T, api *API) {
	testVideo1(t, api)
}

func testVideo1(t *testing.T, api *API) {
	req := api.GrabVideo("uOXLKPCs54c")

	res, err := net.Client.Do(req)
	if err != nil { assert.FailNow(t, err.Error()) }

	var v data.Video
	v.ID = "uOXLKPCs54c"
	recm, err := api.ParseVideo(&v, res)
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
	assert.Equal(t, "Webdriver Torso", v.Uploader)
	assert.Equal(t, "tmpFmHnGe", v.Title)
	assert.Equal(t, "Creative Commons Attribution license (reuse allowed)", v.License)
	assert.Equal(t, "People & Blogs", v.Genre)
	assert.Equal(t, data.VisibilityPublic, v.Visibility)
	assert.False(t, v.NoComments, "no comments")
	assert.False(t, v.NoRatings, "no ratings")
	assert.False(t, v.NoEmbed, "no embed")
	assert.False(t, v.ProductPlacement, "product placement")
	assert.NotZero(t, v.Views, "views")
}
