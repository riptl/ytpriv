package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
)

var videoDetailCmd = cobra.Command{
	Use: "detail <video ID> [file]",
	Short: "Get details about a video",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doVideoDetail),
}

func doVideoDetail(_ *cobra.Command, args []string) error {
	videoID := args[0]

	videoID, err := api.GetVideoID(videoID)
	if err != nil { return err }

	videoReq := apis.Main.GrabVideo(videoID)

	res, err := net.Client.Do(videoReq)
	if err != nil { return err }

	var v data.Video
	v.ID = videoID
	related, err := apis.Main.ParseVideo(&v, res)
	if err != nil { return err }

	bytesMain, err := json.MarshalIndent(&v, "", "\t")
	if err != nil { return err }

	fmt.Println(string(bytesMain))
	fmt.Println()

	if len(related) > 0 {
		bytesRelated, err := json.Marshal(related)
		if err != nil { return err }

		fmt.Println("Related URLs:")
		fmt.Println(string(bytesRelated))
	}

	return nil
}
