package browseajax

import (
	"net/http"
	"github.com/terorie/yt-mango/common"
	"errors"
	"io/ioutil"
	"github.com/valyala/fastjson"
)

const mainURL = "https://www.youtube.com/browse_ajax?ctoken="

func GrabPage(channelID string, page uint) (*fastjson.Value, error) {
	// Generate page URL
	token := GenerateToken(channelID, uint64(page))
	url := mainURL + token

	// Prepare request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil { return nil, err }
	req.Header.Add("X-YouTube-Client-Name", "1")
	req.Header.Add("X-YouTube-Client-Version", "2.20180726")

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
	root, err := p.ParseBytes(buf)
	if err != nil { return nil, err }

	return root, nil
}
