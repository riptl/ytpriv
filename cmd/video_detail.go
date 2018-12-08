package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var videoDetailCmd = cobra.Command{
	Use: "detail <video> [file]",
	Short: "Get details about a video",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doVideoDetail),
}

func doVideoDetail(_ *cobra.Command, args []string) error {
	videoID := args[0]

	videoID, err := api.GetVideoID(videoID)
	if err != nil { return err }

	videoReq := apis.Main.GrabVideo(videoID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = net.Client.Do(videoReq, res)
	if err != nil { return err }

	var v data.Video
	v.ID = videoID
	err = apis.Main.ParseVideo(&v, res)
	if err != nil { return err }

	bytesMain, err := json.MarshalIndent(&v, "", "\t")
	if err != nil { return err }

	fmt.Println(string(bytesMain))
	fmt.Println()

	return nil
}
