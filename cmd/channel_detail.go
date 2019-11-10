package cmd

import (
	"github.com/spf13/cobra"
)

var channelDetailCmd = cobra.Command{
	Use: "detail <channel ID>",
	Short: "Get detail about a channel",
	Args: cobra.ExactArgs(1),
	Run: cmdFunc(doChannelDetail),
}

func doChannelDetail(_ *cobra.Command, args []string) error {
	panic("not implemented")
	/*channelID := args[0]

	channelID, err := api.GetChannelID(channelID)
	if err != nil { return err }

	channelReq := api.GrabChannel(channelID)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	err = net.Client.Do(channelReq, res)
	if err != nil { return err }

	var c data.Channel
	api.ParseChannel(&c, res)

	bytes, err := json.MarshalIndent(&c, "", "\t")
	if err != nil { return err }
	fmt.Println(string(bytes))

	return nil*/
}
