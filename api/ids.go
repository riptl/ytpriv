package api

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// FIXME: API package should be abstract, no utility code in here

var matchChannelID = regexp.MustCompile("^([\\w\\-]|(%3[dD]))+$")
var matchVideoID = regexp.MustCompile("^[\\w\\-]+$")

// Input: Channel ID or link to YT channel page
// Output: Channel ID or "" on error
func GetChannelID(chanURL string) (string, error) {
	if !matchChannelID.MatchString(chanURL) {
		// Check if youtube.com domain
		_url, err := url.Parse(chanURL)
		if err != nil || (_url.Host != "www.youtube.com" && _url.Host != "youtube.com") {
			return "", fmt.Errorf("not a channel ID: %s", chanURL)
		}

		// Check if old /user/ URL
		if strings.HasPrefix(_url.Path, "/user/") {
			// TODO Implement extraction of channel ID
			return "", fmt.Errorf("New /channel/ link is required!\n" +
				"The old /user/ links do not work: %s", chanURL)
		}

		// Remove /channel/ path
		channelID := strings.TrimPrefix(_url.Path, "/channel/")
		if len(channelID) == len(_url.Path) {
			// No such prefix to be removed
			return "", fmt.Errorf("not a channel ID: %s", channelID)
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

func GetVideoID(vidURL string) (string, error) {
	extractors := []func(string)(string, error) {
		getVideoIdFromUrl,
		getVideoIdFromShortUrl,
		getVideoIdFromUrlVariant,
	}

	for _, f := range extractors {
		vidID, err := f(vidURL)
		// Explicit error:
		// Extractor right but malformed input
		if err != nil { return "", err }
		// Success
		if vidID != "" {
			return vidID, nil
		}
		// Wrong extractor, continue
	}

	// No extractor worked, check if already ID
	if matchVideoID.MatchString(vidURL) {
		return vidURL, nil
	} else {
		// No match
		return "", fmt.Errorf("could not extract video ID: %s", vidURL)
	}
}

// Extracts the ID from watch?v links
func getVideoIdFromUrl(vidURL string) (string, error) {
	// Not a watch URL
	if !strings.Contains(vidURL, "youtube.com/watch") {
		return "", nil
	}

	// Parse URL
	_url, err := url.Parse(vidURL)
	if err != nil { return "", err }

	// Extract from query
	vidID := _url.Query().Get("v")
	if vidID == "" { return "", fmt.Errorf("invalid input: %s", vidURL) }

	return vidID, nil
}

// Extracts the ID from links like "youtu.be/<id>"
func getVideoIdFromShortUrl(vidURL string) (string, error) {
	if !strings.Contains(vidURL, "youtu.be/") {
		return "", nil
	}

	_url, err := url.Parse(vidURL)
	if err != nil { return "", err }

	return strings.TrimPrefix(_url.Path, "/"), nil
}

// Extracts the ID from links like "www.youtube.com/v/<id>?params=…"
func getVideoIdFromUrlVariant(vidURL string) (string, error) {
	urlV := strings.Contains(vidURL, "youtube.com/v/")
	urlEmbed := strings.Contains(vidURL, "youtube.com/embed/")

	var pathPrefix string
	switch {
		case urlV: pathPrefix = "/v/"
		case urlEmbed: pathPrefix = "/embed/"
		default: return "", nil
	}

	// Parse URL
	_url, err := url.Parse(vidURL)
	if err != nil { return "", err }

	// Extract ID by splitting prefix
	vidID := strings.TrimPrefix(_url.Path, pathPrefix)
	if len(_url.Path) == len(vidID) { return "", fmt.Errorf("invalid input: %s", vidURL) }

	// Split at first "/"
	vidID = strings.SplitN(vidID, "/", 1)[0]

	return vidID, nil
}
