package cmd

import (
	"github.com/spf13/cobra"
	"regexp"
)

var force bool
var offset uint32

var Channel = cobra.Command{
	Use: "channel",
	Short: "Get information about a channel",
}

var matchChannelID = regexp.MustCompile("^([\\w\\-]|(%3[dD]))+$")

func init() {
	channelDumpCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite the output file if it already exists")
	channelDumpCmd.Flags().Uint32Var(&offset, "page-offset", 1, "Start getting videos at this page. (A page is usually 30 videos)")
	Channel.AddCommand(&channelDumpCmd)
}
