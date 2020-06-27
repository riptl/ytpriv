package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	csv "github.com/terorie/go-quotecsv"
	"github.com/valyala/fastjson"
)

var nullVal = []byte("NULL")

func main() {
	scn := bufio.NewScanner(os.Stdin)
	wr := csv.NewWriter(os.Stdout)
	defer wr.Flush()
	buf := make([]byte, 0, 1024 * 1024)
	scn.Buffer(buf, 128 * 1024 * 1024)
	var p fastjson.Parser
	i := 0
	for scn.Scan() {
		i++
		obj, err := p.ParseBytes(scn.Bytes())
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Failed to parse line %d: %s\n", i, err)
		}
		id := obj.GetStringBytes("id")
		if len(id) == 0 {
			continue
		}
		vid := obj.GetStringBytes("video_id")
		if len(vid) == 0 {
			continue
		}
		aid := obj.GetStringBytes("author_id")
		if len(aid) == 0 {
			continue
		}
		pid := obj.GetStringBytes("parent_id")
		if len(pid) == 0 {
			pid = nullVal
		}
		cwat := obj.GetInt64("crawled_at")
		likes := obj.GetInt64("likes")
		replies := obj.GetInt64("replies")
		aname := obj.GetStringBytes("author")
		cobj := obj.Get("content")
		// Comment text
		for _, line := range cobj.GetArray() {
			if nav := line.Get("navigationEndpoint"); nav != nil {
				nav.Del("clickTrackingParams")
				nav.Del("commandMetadata")
			}
		}

		var content []byte
		if cobj != nil {
			content = cobj.MarshalTo(nil)
		}

		_ = wr.Write([]csv.Column{
			{Value: string(id), Quoted: true},
			{Value: string(vid), Quoted: true},
			{Value: string(aid), Quoted: true},
			{Value: string(pid)},
			{Value: strconv.FormatInt(cwat, 10)},
			{Value: strconv.FormatInt(likes, 10)},
			{Value: strconv.FormatInt(replies, 10)},
			{Value: string(aname), Quoted: true},
			{Value: string(content), Quoted: true},
		})
	}
	if err := scn.Err(); err != nil {
		panic(err)
	}
}
