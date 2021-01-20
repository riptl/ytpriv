package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var playlistVideos = cobra.Command{
	Use:   "videos <playlist ID>",
	Short: "Dump the video IDs in a playlist",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(playlistVideosCmd),
}

func playlistVideosCmd(_ *cobra.Command, args []string) error {
	id := args[0]
	playlist, err := client.RequestPlaylist(id).Do()
	if err != nil {
		return err
	}
	for _, video := range playlist.Videos {
		fmt.Println(video.ID)
	}
	return nil
}
