package net

import "net/http"

// Custom headers
type transport struct{}

// Important:
// - Set header "Accept-Language: en-US" or else parser might break
// - Set header "User-Agent: youtube-mango/1.0"
func (t transport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("Accept-Language", "en-US")
	r.Header.Add("User-Agent", "youtube-mango/0.1")
	return http.DefaultTransport.RoundTrip(r)
}

var Client = http.Client{Transport: transport{}}
