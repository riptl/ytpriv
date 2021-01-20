package main

import (
	"os"

	"github.com/spf13/cobra"
	yt "github.com/terorie/ytwrk"
	"github.com/valyala/fasthttp"
)

var videoRawCmd = cobra.Command{
	Use:   "raw <video>",
	Short: "Dump raw JSON of video",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doVideoRaw),
}

func doVideoRaw(_ *cobra.Command, args []string) error {
	videoID := args[0]
	videoID, err := yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}
	req := client.RequestVideo(videoID)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := client.HTTP.Do(req.Request, res); err != nil {
		return err
	}
	return res.BodyWriteTo(os.Stdout)
}
