package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var playlistVideos = cobra.Command{
	Use:   "videos <playlist ID>",
	Short: "Dump the video IDs in a playlist",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(playlistVideosCmd),
}

func playlistVideosCmd(_ *cobra.Command, args []string) error {
	id := args[0]
	playlist, err := client.RequestPlaylistStart(id).Do()
	if err != nil {
		return err
	}
	for _, video := range playlist.Page.Videos {
		fmt.Println(video.ID)
	}
	return nil
}

var playlistVideosPageCmd = cobra.Command{
	Use: "videos_page <playlist ID or continuation>",
	Short: "Get page of videos of channel",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doPlaylistVideosPage),
}

func init() {
	flags := playlistVideosPageCmd.Flags()
	flags.Bool("raw", false, "Dump raw JSON")

	playlistCmd.AddCommand(&playlistVideosPageCmd)
}

func doPlaylistVideosPage(c *cobra.Command, args []string) error {
	arg := args[0]
	raw, err := c.Flags().GetBool("raw")
	if err != nil {
		panic(err)
	}
	if len(arg) == 34 {
		// Playlist ID
		req := client.RequestPlaylistStart(arg)
		if !raw {
			playlist, err := req.Do()
			if err != nil {
				return err
			}
			bytesMain, err := json.MarshalIndent(playlist, "", "\t")
			if err != nil {
				return err
			}
			fmt.Println(string(bytesMain))
			return nil
		} else {
			res := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(res)
			if err := client.HTTP.Do(req.Request, res); err != nil {
				return err
			}
			return res.BodyWriteTo(os.Stdout)
		}
	} else {
		// Continuation
		req := client.RequestPlaylistPage(arg)
		if !raw {
			playlist, err := req.Do()
			if err != nil {
				return err
			}
			bytesMain, err := json.MarshalIndent(playlist, "", "\t")
			if err != nil {
				return err
			}
			fmt.Println(string(bytesMain))
			return nil
		} else {
			res := fasthttp.AcquireResponse()
			defer fasthttp.ReleaseResponse(res)
			if err := client.HTTP.Do(req.Request, res); err != nil {
				return err
			}
			return res.BodyWriteTo(os.Stdout)
		}
	}
}
