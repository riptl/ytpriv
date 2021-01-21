package yt

import (
	"encoding/xml"
	"errors"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

// Client accesses the YouTube private API.
type Client struct {
	HTTP *fasthttp.Client
}

func NewClient() *Client {
	return &Client{HTTP: &fasthttp.Client{
		Name:         "ytwrk/testing",
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}}
}

var ErrVideoUnavailable = errors.New("video unavailable")
var MissingData = errors.New("missing data")
var ServerError = errors.New("server error")

func (c *Client) GetVideoSubtitleList(videoID string) (tracks *XMLSubTrackList, err error) {
	const subtitleURL = "https://video.google.com/timedtext?type=list&v="
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.SetRequestURI(subtitleURL + videoID)
	setXMLHeaders(&req.Header)
	res := fasthttp.AcquireResponse()
	err = c.HTTP.Do(req, res)
	if err != nil {
		return
	}
	if res.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", res.StatusCode())
	}
	tracks = new(XMLSubTrackList)
	err = xml.Unmarshal(res.Body(), tracks)
	return
}

func setHeaders(h *fasthttp.RequestHeader) {
	h.Add("Accept-Language", "en-US")
	h.Add("Host", "www.youtube.com")
	h.Add("X-Origin", "https://www.youtube.com")
	h.Add("X-YouTube-Client-Name", "1")
	h.Add("X-YouTube-Client-Version", "2.20210119.08.00")
}

func setXMLHeaders(h *fasthttp.RequestHeader) {
	h.Add("Accept-Language", "en-US")
	h.Add("Host", "www.youtube.com")
}

type XMLSubTrackList struct {
	Tracks []struct {
		LangCode string `xml:"lang_code,attr"`
		Lang     string `xml:"lang_translated,attr"`
	} `xml:"track"`
}

var defaultYTIContext = ytiContext{Client: ytiClient{
	IsInternal:                  true,
	ClientName:                  "WEB",
	ClientVersion:               "2.20210119.08.00",
	InternalClientExperimentIDs: []int{44496012},
	Platform:                    "DESKTOP",
}}

type ytiRequest struct {
	BrowseID     string      `json:"browseId,omitempty"`
	Continuation string      `json:"continuation,omitempty"`
	Context      *ytiContext `json:"context"`
	Params       string      `json:"params,omitempty"`
}

type ytiContext struct {
	Client ytiClient `json:"client"`
}

type ytiClient struct {
	IsInternal                  bool   `json:"isInternal"`
	ClientName                  string `json:"clientName"`
	ClientVersion               string `json:"clientVersion"`
	InternalClientExperimentIDs []int  `json:"internalClientExperimentIds"`
	Platform                    string `json:"platform"`
}
