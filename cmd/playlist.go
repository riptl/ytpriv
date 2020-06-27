package cmd

import "github.com/spf13/cobra"

var Playlist = cobra.Command{
	Use:   "playlist",
	Short: "Get information about playlists",
}

func init() {
	Playlist.AddCommand(
		&playlistVideos,
	)
}
