package apijson

import (
	"github.com/terorie/yt-mango/data"
	"net/http"
	"github.com/terorie/yt-mango/common"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"errors"
)

const videoURL = "https://www.youtube.com/watch?pbj=1&v="
const channelURL = "https://www.youtube.com/browse_ajax?ctoken="

func GrabVideo(v *data.Video) (root *fastjson.Value, err error) {
	// Prepare request
	req, err := http.NewRequest("GET", videoURL+ v.ID, nil)
	if err != nil { return nil, err }
	setHeaders(&req.Header)

	// Send request
	res, err := common.Client.Do(req)
	if err != nil { return }

	// Download response
	body, err := ioutil.ReadAll(res.Body)
	if err != nil { return }

	// Parse JSON
	var p fastjson.Parser
	root, err = p.ParseBytes(body)
	if err != nil { return }

	return
}

func GrabChannelPage(channelID string, page uint) (root *fastjson.Value, err error) {
	// Generate page URL
	token := GenChannelPageToken(channelID, uint64(page))
	url := channelURL + token

	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return nil, err }
	setHeaders(&req.Header)

	// Send request
	res, err := common.Client.Do(req)
	if err != nil { return nil, err }
	if res.StatusCode == 500 {
		defer res.Body.Close()
		buf, _ := ioutil.ReadAll(res.Body)
		println(string(buf))
	}
	if res.StatusCode != 200 { return nil, errors.New("HTTP failure") }

	// Download response
	defer res.Body.Close()
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil { return nil, err }

	// Parse JSON
	var p fastjson.Parser
	root, err = p.ParseBytes(buf)
	return
}

func setHeaders(h *http.Header) {
	h.Add("Host", "www.youtube.com")
	h.Add("User-Agent", "yt-mango/0.1")
	h.Add("X-YouTube-Client-Name", "1")
	h.Add("X-YouTube-Client-Version", "2.20170707")
}
