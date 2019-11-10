package cmd

import (
	"github.com/spf13/cobra"
)

var videoSubtitlesCmd = cobra.Command{
	Use:   "subtitles",
	Short: "Access video subtitles",
}

func init() {
	videoSubtitlesCmd.AddCommand(&videoSubtitlesDumpCmd)
}

var videoSubtitlesDumpCmd = cobra.Command{
	Use:   "dump",
	Short: "Dump all available subtitles",
	Args:  cobra.ExactArgs(1),
}
