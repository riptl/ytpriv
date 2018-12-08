package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var channelDetailCmd = cobra.Command{
	Use: "detail <channel ID>",
	Short: "Get detail about a channel",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doChannelDetail),
}

func doChannelDetail(_ *cobra.Command, args []string) error {
	channelID := args[0]

	channelID, err := api.GetChannelID(channelID)
	if err != nil { return err }

	channelReq := apis.Main.GrabChannel(channelID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = net.Client.Do(channelReq, res)
	if err != nil { return err }

	var c data.Channel
	apis.Main.ParseChannel(&c, res)

	bytes, err := json.MarshalIndent(&c, "", "\t")
	if err != nil { return err }
	fmt.Println(string(bytes))

	return nil
}
