package apiclassic

import (
	"net/http"
	"errors"
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"github.com/terorie/yt-mango/common"
)

const mainURL = "https://www.youtube.com/watch?has_verified=1&bpctr=6969696969&v="
const subtitleURL = "https://video.google.com/timedtext?type=list&v="

// Grabs a HTML video page and returns the document tree
func GrabVideo(videoID string) (doc *goquery.Document, err error) {
	req, err := http.NewRequest("GET", mainURL + videoID, nil)
	if err != nil { return }
	setHeaders(&req.Header)

	res, err := common.Client.Do(req)
	if err != nil { return }
	if res.StatusCode != 200 { return nil, errors.New("HTTP failure") }

	defer res.Body.Close()
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil { return nil, err }

	return
}

// Grabs and parses a subtitle list
func GrabSubtitleList(videoID string) (tracks *XMLSubTrackList, err error) {
	req, err := http.NewRequest("GET", subtitleURL + videoID, nil)
	if err != nil { return }
	setHeaders(&req.Header)

	res, err := common.Client.Do(req)
	if err != nil { return }
	if res.StatusCode != 200 { return nil, errors.New("HTTP failure") }

	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)

	tracks = new(XMLSubTrackList)
	err = decoder.Decode(tracks)
	return
}

func setHeaders(h *http.Header) {
	h.Add("Host", "www.youtube.com")
	h.Add("User-Agent", "yt-mango/0.1")
}
