package classic

import (
	"net/http"
	"errors"
	"encoding/xml"
	"github.com/PuerkitoBio/goquery"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/common"
)

const mainURL = "https://www.youtube.com/watch?has_verified=1&bpctr=6969696969&v="
const subtitleURL = "https://video.google.com/timedtext?type=list&v="

// Grabs a HTML video page and returns the document tree
func grab(v *data.Video) (doc *goquery.Document, err error) {
	req, err := http.NewRequest("GET", mainURL + v.ID, nil)
	if err != nil { return }
	requestHeader(&req.Header)

	res, err := common.Client.Do(req)
	if err != nil { return }
	if res.StatusCode != 200 { return nil, errors.New("HTTP failure") }

	defer res.Body.Close()
	doc, err = goquery.NewDocumentFromReader(res.Body)
	if err != nil { return nil, err }

	return
}

// Grabs and parses a subtitle list
func grabSubtitleList(v *data.Video) (err error) {
	req, err := http.NewRequest("GET", subtitleURL + v.ID, nil)
	if err != nil { return err }

	res, err := client.Do(req)
	if err != nil { return err }
	if res.StatusCode != 200 { return errors.New("HTTP failure") }

	defer res.Body.Close()
	decoder := xml.NewDecoder(res.Body)

	var tracks XMLSubTrackList
	err = decoder.Decode(&tracks)
	if err != nil { return err }

	for _, track := range tracks.Tracks {
		v.Subtitles = append(v.Subtitles, track.LangCode)
	}

	return
}
