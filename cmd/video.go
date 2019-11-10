package cmd

import "github.com/spf13/cobra"

var Video = cobra.Command{
	Use: "video",
	Short: "Get information about a video",
}

func init() {
	Video.AddCommand(
		&videoCommentsCmd,
		&videoDetailCmd,
		&videoDumpCmd,
		&videoDumpRawCmd,
	)
}
