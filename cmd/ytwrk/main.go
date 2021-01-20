package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := Root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var channelCmd = cobra.Command{
	Use:   "channel",
	Short: "Get information about a channel",
}

var playlistCmd = cobra.Command{
	Use:   "playlist",
	Short: "Get information about playlists",
}

func init() {
	playlistCmd.AddCommand(
		&playlistVideos,
	)
}

var videoCmd = cobra.Command{
	Use:   "video",
	Short: "Get information about a video",
}

func init() {
	videoCmd.AddCommand(
		&videoCommentsCmd,
		&videoDetailCmd,
		&videoLiveCmd,
		&videoRawCmd,
	)
}
