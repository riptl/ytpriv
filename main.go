/* youtube-ma for MongoDB
 *
 * Based on https://github.com/CorentinB/youtube-ma */

package main

import (
	"encoding/json"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/classic"
)

func main() {
	v := data.Video{ID: "kj9mFK62c6E"}

	err := classic.Get(&v)
	if err != nil { panic(err) }

	jsn, err := json.MarshalIndent(v, "", "\t")
	if err != nil { panic(err) }

	println(string(jsn))
}
