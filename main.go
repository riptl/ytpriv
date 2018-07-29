/* youtube-ma for MongoDB
 *
 * Based on https://github.com/CorentinB/youtube-ma */

package main

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
)

const Version = "v0.1 -- dev"

func printVersion(_ *cobra.Command, _ []string) {
	fmt.Println("YT-Mango archiver", Version)
}

func main() {
	rootCmd := cobra.Command{
		Use:   "yt-mango",
		Short: "YT-Mango is a scalable video metadata archiver",
		Long: "YT-Mango is a scalable video metadata archiving utility\n" +
			"written by terorie for https://the-eye.eu/",
	}

	versionCmd := cobra.Command{
		Use: "version",
		Short: "Get the version number of yt-mango",
		Run: printVersion,
	}

	rootCmd.AddCommand(&versionCmd)
	rootCmd.AddCommand(&channelCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
