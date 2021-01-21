package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	yt "github.com/terorie/ytwrk"
	"github.com/terorie/ytwrk/types"
	"github.com/valyala/fasthttp"
)

var channelOverviewCmd = cobra.Command{
	Use: "overview <channel>",
	Short: "Get overview of channel",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doChannelOverview),
}

func init() {
	flags := channelOverviewCmd.Flags()
	flags.Bool("raw", false, "Dump raw JSON")

	channelCmd.AddCommand(&channelOverviewCmd)
}

func doChannelOverview(c *cobra.Command, args []string) error {
	channelID := args[0]
	channelID, err := yt.ExtractChannelID(channelID)
	if err != nil {
		return err
	}
	req := client.RequestChannelOverview(channelID)
	raw, err := c.Flags().GetBool("raw")
	if err != nil {
		panic(err)
	}
	if !raw {
		page, err := req.Do()
		if err != nil {
			return err
		}
		bytesMain, err := json.MarshalIndent(page, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytesMain))
		fmt.Println()
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

var channelVideosCmd = cobra.Command{
	Use: "videos <channel ID>",
	Short: "Get full list of videos of channel",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doChannelVideos),
}

func init() {
	channelCmd.AddCommand(&channelVideosCmd)
}

func doChannelVideos(_ *cobra.Command, args []string) error {
	channelID := args[0]
	channelID, err := yt.ExtractChannelID(channelID)
	if err != nil {
		return fmt.Errorf("failed to extract channel ID: %w", err)
	}
	page, err := client.RequestChannelVideosStart(channelID).Do()
	if err != nil {
		return fmt.Errorf("failed to get page 0: %w", err)
	}
	for _, video := range page.Videos {
		res, err := json.Marshal(video)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(res))
	}
	pageNum := 1
	for page.Continuation != "" {
		page, err = client.RequestChannelVideosPage(page.Continuation).Do()
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

var channelVideosPageCmd = cobra.Command{
	Use:   "videos_page <channel ID>",
	Short: "Get videos page of channel",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doChannelVideosPage),
}

func init() {
	flags := channelVideosPageCmd.Flags()
	flags.Bool("raw", false, "Dump raw JSON")

	channelCmd.AddCommand(&channelVideosPageCmd)
}

func doChannelVideosPage(c *cobra.Command, args []string) error {
	arg := args[0]
	raw, err := c.Flags().GetBool("raw")
	if err != nil {
		panic(err)
	}
	// TODO This is a bit ugly
	var req interface{
		Do() (*types.ChannelVideosPage, error)
		GetRequest() *fasthttp.Request
	}
	if len(arg) == 24 {
		// Channel ID
		req = client.RequestChannelVideosStart(arg)
	} else {
		req = client.RequestChannelVideosPage(arg)
	}
	if !raw {
		page, err := req.Do()
		if err != nil {
			return err
		}
		bytesMain, err := json.MarshalIndent(page, "", "\t")
		if err != nil {
			return err
		}
		fmt.Println(string(bytesMain))
		fmt.Println()
		return nil
	} else {
		res := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(res)
		if err := client.HTTP.Do(req.GetRequest(), res); err != nil {
			return err
		}
		return res.BodyWriteTo(os.Stdout)
	}
}
