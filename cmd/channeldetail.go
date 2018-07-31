package cmd

import (
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"os"
	"log"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"fmt"
	"encoding/json"
)

var channelDetailCmd = cobra.Command{
	Use: "detail <channel ID>",
	Short: "Get detail about a channel",
	Args: cobra.ExactArgs(1),
	Run: doChannelDetail,
}

func doChannelDetail(_ *cobra.Command, args []string) {
	channelID := args[0]

	channelID = api.GetChannelID(channelID)
	if channelID == "" {
		os.Exit(1)
	}

	channelReq := api.Main.GrabChannel(channelID)

	res, err := net.Client.Do(channelReq)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	var c data.Channel
	api.Main.ParseChannel(&c, res)

	bytes, err := json.MarshalIndent(&c, "", "\t")
	if err != nil { panic(err) }
	fmt.Println(string(bytes))
}
