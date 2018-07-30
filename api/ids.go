package api

import (
	"regexp"
	"os"
	"strings"
	"log"
	"net/url"
)

var matchChannelID = regexp.MustCompile("^([\\w\\-]|(%3[dD]))+$")

func GetChannelID(chanURL string) (string, error) {
	if !matchChannelID.MatchString(chanURL) {
		// Check if youtube.com domain
		_url, err := url.Parse(chanURL)
		if err != nil || (_url.Host != "www.youtube.com" && _url.Host != "youtube.com") {
			log.Fatal("Not a channel ID:", chanURL)
			os.Exit(1)
		}

		// Check if old /user/ URL
		if strings.HasPrefix(_url.Path, "/user/") {
			// TODO Implement extraction of channel ID
			log.Fatal("New /channel/ link is required!\n" +
				"The old /user/ links do not work.")
			os.Exit(1)
		}

		// Remove /channel/ path
		channelID := strings.TrimPrefix(_url.Path, "/channel/")
		if len(channelID) == len(_url.Path) {
			// No such prefix to be removed
			log.Fatal("Not a channel ID:", channelID)
			os.Exit(1)
		}

		// Remove rest of path from channel ID
		slashIndex := strings.IndexRune(channelID, '/')
		if slashIndex != -1 {
			channelID = channelID[:slashIndex]
		}

		return channelID, nil
	} else {
		// It's already a channel ID
		return chanURL, nil
	}
}
