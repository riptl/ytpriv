package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/valyala/fastjson"
)

func main() {
	fmt.Println("id\tvideo_id\tauthor_id\tcrawled_at\tlikes\tauthor\tcontent")
	dec := json.NewDecoder(os.Stdin)
	i := 0
	type line struct {
		Id string `json:"id"`
		Vid string `json:"video_id"`
		Aid string `json:"author_id"`
		Cwat json.Number `json:"crawled_at"`
		Likes json.Number `json:"likes"`
		Aname string `json:"author"`
		Content json.RawMessage `json:"content"`
	}

	var p fastjson.Parser
	for dec.More() {
		i++
		var l line
		err := dec.Decode(&l)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to parse line %d: %s\n", i, err)
		}
		cobj, _ := p.ParseBytes(l.Content)
		for _, cem := range cobj.GetArray() {
			if nav := cem.Get("navigationEndpoint"); nav != nil {
				nav.Del("clickTrackingParams")
				nav.Del("commandMetadata")
			}
		}

		content, err := json.Marshal(l.Content)
		if err != nil { panic(err) }

		fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", l.Id, l.Vid, l.Aid, l.Cwat, l.Likes, l.Aname, content)
	}
}
