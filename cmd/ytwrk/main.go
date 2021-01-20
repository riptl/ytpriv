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

var force bool

var Channel = cobra.Command{
	Use:   "channel",
	Short: "Get information about a channel",
}

func init() {
	channelDumpCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite the output file if it already exists")
	Channel.AddCommand(
		&channelDumpRawCmd,
		&channelDumpCmd,
	)
}

var Playlist = cobra.Command{
	Use:   "playlist",
	Short: "Get information about playlists",
}

func init() {
	Playlist.AddCommand(
		&playlistVideos,
	)
}

var Video = cobra.Command{
	Use:   "video",
	Short: "Get information about a video",
}

func init() {
	Video.AddCommand(
		&videoCommentsCmd,
		&videoDetailCmd,
		&videoDumpCmd,
		&videoDumpRawCmd,
		&videoParseRawCmd,
		&VideoLive,
	)
}
