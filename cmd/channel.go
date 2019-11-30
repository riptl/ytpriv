package cmd

import "github.com/spf13/cobra"

var force bool

var Channel = cobra.Command{
	Use: "channel",
	Short: "Get information about a channel",
}

func init() {
	channelDumpCmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite the output file if it already exists")
	Channel.AddCommand(
		&channelDumpRawCmd,
		&channelDumpCmd,
	)
}
