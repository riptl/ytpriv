package cmd

import (
	"os"
	"encoding/json"
	"fmt"
	"log"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
)

var videoDetailCmd = cobra.Command{
	Use: "detail <video ID> [file]",
	Short: "Get details about a video",
	Args: cobra.ExactArgs(1),
	Run: doVideoDetail,
}

func doVideoDetail(_ *cobra.Command, args []string) {
	videoID := args[0]

	videoID, err := api.GetVideoID(videoID)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	videoReq := api.Main.GrabVideo(videoID)

	res, err := net.Client.Do(videoReq)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var v data.Video
	v.ID = videoID
	api.Main.ParseVideo(&v, res)

	bytes, err := json.MarshalIndent(&v, "", "\t")
	if err != nil { panic(err) }
	fmt.Println(string(bytes))
}
