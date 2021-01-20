package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	yt "github.com/terorie/ytwrk"
)

var videoDetailCmd = cobra.Command{
	Use:   "detail <video>",
	Short: "Get details about a video",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doVideoDetail),
}

func doVideoDetail(_ *cobra.Command, args []string) error {
	videoID := args[0]
	videoID, err := yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}
	video, err := client.RequestVideo(videoID).Do()
	if err != nil {
		return err
	}
	bytesMain, err := json.MarshalIndent(video, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(bytesMain))
	fmt.Println()
	return nil
}
