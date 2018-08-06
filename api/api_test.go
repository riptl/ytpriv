package api

import (
	"testing"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"github.com/stretchr/testify/assert"
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
	assert.Nil(t, err, err)

	var v data.Video
	recm, err := api.ParseVideo(&v, res)
	assert.NotZero(t, len(recm), "No recommendations")

	assert.Equal(t, "tmpFmHnGe", v.Title)
	assert.Equal(t, "People & Blogs", v.Genre)
}
