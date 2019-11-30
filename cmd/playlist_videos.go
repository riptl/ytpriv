package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/data"
	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

var playlistVideos = cobra.Command{
	Use: "videos <playlist ID>",
	Short: "Dump the video IDs in a playlist",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(playlistVideosCmd),
}

func playlistVideosCmd(_ *cobra.Command, args []string) error {
	id := args[0]
	req := api.GrabPlaylist(id)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := net.Client.Do(req, res); err != nil {
		return err
	}
	var l data.Playlist
	if err := api.ParsePlaylist(&l, res); err != nil {
		return err
	}
	for _, video := range l.Videos {
		fmt.Println(video.ID)
	}
	return nil
}
