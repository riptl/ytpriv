package cmd

import "github.com/spf13/cobra"

var videoDetailCmd = cobra.Command{
	Use: "detail <video ID> [file]",
	Short: "Get details about a video",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init()  {
}
