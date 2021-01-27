package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/valyala/fasthttp"
)

var playlistCmd = cobra.Command{
	Use:   "playlist",
	Short: "Scrape a playlist",
}

func init() {
	rootCmd.AddCommand(&playlistCmd)
}

var playlistVideos = cobra.Command{
	Use:   "videos <playlist ID>",
	Short: "Get full list of videos in playlist",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doPlaylistVideos),
}

func init() {
	playlistCmd.AddCommand(&playlistVideos)
}

func doPlaylistVideos(_ *cobra.Command, args []string) error {
	playlistID := args[0]
	// TODO Extract ID if video or playlist URL
	startPage, err := client.RequestPlaylistStart(playlistID).Do()
	if err != nil {
		return fmt.Errorf("failed to get page 0: %w", err)
	}
	page := &startPage.Page
	for _, video := range page.Videos {
		res, err := json.Marshal(video)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(res))
	}
	pageNum := 1
	for page.Continuation != "" {
		page, err = client.RequestPlaylistPage(page.Continuation).Do()
		if err != nil {
			return fmt.Errorf("failed to get page %d: %w", pageNum, err)
		}
		for _, video := range page.Videos {
			res, err := json.Marshal(video)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(res))
		}
		pageNum++
	}
	return nil
}

var playlistVideosPageCmd = cobra.Command{
	Use:   "videos_page <playlist ID or continuation>",
	Short: "Get page of videos of playlist",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doPlaylistVideosPage),
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
