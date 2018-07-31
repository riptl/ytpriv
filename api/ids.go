package api

import (
	"regexp"
	"strings"
	"log"
	"net/url"
)

// FIXME: API package should be abstract, no utility code in here

var matchChannelID = regexp.MustCompile("^([\\w\\-]|(%3[dD]))+$")
var matchVideoID = regexp.MustCompile("^[\\w\\-]+$")

// Input: Channel ID or link to YT channel page
// Output: Channel ID or "" on error
func GetChannelID(chanURL string) string {
	if !matchChannelID.MatchString(chanURL) {
		// Check if youtube.com domain
		_url, err := url.Parse(chanURL)
		if err != nil || (_url.Host != "www.youtube.com" && _url.Host != "youtube.com") {
			log.Fatal("Not a channel ID:", chanURL)
			return ""
		}

		// Check if old /user/ URL
		if strings.HasPrefix(_url.Path, "/user/") {
			// TODO Implement extraction of channel ID
			log.Print("New /channel/ link is required!\n" +
				"The old /user/ links do not work:", chanURL)
			return ""
		}

		// Remove /channel/ path
		channelID := strings.TrimPrefix(_url.Path, "/channel/")
		if len(channelID) == len(_url.Path) {
			// No such prefix to be removed
			log.Print("Not a channel ID:", channelID)
			return ""
		}

		// Remove rest of path from channel ID
		slashIndex := strings.IndexRune(channelID, '/')
		if slashIndex != -1 {
			channelID = channelID[:slashIndex]
		}

		return channelID
	} else {
		// It's already a channel ID
		return chanURL
	}
}

func GetVideoID(vidURL string) string {
	if !matchVideoID.MatchString(vidURL) {
		// Check if youtube.com domain
		_url, err := url.Parse(vidURL)
		if err != nil || (_url.Host != "www.youtube.com" && _url.Host != "youtube.com") {
			log.Fatal("Not a video ID:", vidURL)
			return ""
		}

		// TODO Support other URLs (/v or /embed)

		// Check if watch path
		if !strings.HasPrefix(_url.Path, "/watch") {
			log.Fatal("Not a watch URL:", vidURL)
			return ""
		}

		// Parse query string
		query := _url.Query()
		videoID := query.Get("v")
		if videoID == "" {
			log.Fatal("Invalid watch URL:", vidURL)
			return ""
		}

		return videoID
	} else {
		return vidURL
	}
}
