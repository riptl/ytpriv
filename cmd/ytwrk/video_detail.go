package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	yt "github.com/terorie/ytpriv"
	"github.com/valyala/fasthttp"
)

var videoCmd = cobra.Command{
	Use:   "video",
	Short: "Scrape a video",
}

func init() {
	rootCmd.AddCommand(&videoCmd)
}

var videoDetailCmd = cobra.Command{
	Use:   "detail <video>",
	Short: "Get details about a video",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doVideoDetail),
}

func init() {
	flags := videoDetailCmd.Flags()
	flags.Bool("raw", false, "Dump raw JSON")

	videoCmd.AddCommand(&videoDetailCmd)
}

func doVideoDetail(c *cobra.Command, args []string) error {
	videoID := args[0]
	videoID, err := yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}
	req := client.RequestVideo(videoID)
	raw, err := c.Flags().GetBool("raw")
	if err != nil {
		panic(err)
	}
	if !raw {
		video, err := req.Do()
		if err != nil {
			return err
		}
		bytesMain, err := json.MarshalIndent(video, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytesMain))
		fmt.Println()
	} else {
		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)
		if err := client.HTTP.Do(req.Request, res); err != nil {
			return err
		}
		return res.BodyWriteTo(os.Stdout)
	}
	return nil
}
