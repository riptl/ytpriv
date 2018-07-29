// yt-mango: YT video metadata archiving utility
// Copyright (C) 2018  terorie

package main

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"github.com/terorie/yt-mango/cmd"
	"log"
	"github.com/terorie/yt-mango/api"
)

const Version = "v0.1 -- dev"

func printVersion(_ *cobra.Command, _ []string) {
	fmt.Println("YT-Mango archiver", Version)
}

func main() {
	// All diagnostics (logging) should go to stderr
	log.SetOutput(os.Stderr)

	var printVersion bool
	var forceAPI string

	rootCmd := cobra.Command{
		Use:   "yt-mango",
		Short: "YT-Mango is a scalable video metadata archiver",
		Long: "YT-Mango is a scalable video metadata archiving utility\n" +
			"written by terorie for https://the-eye.eu/",
		PreRun: func(cmd *cobra.Command, args []string) {
			if printVersion {
				fmt.Println(Version)
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			switch forceAPI {
			case "": api.DefaultAPI = &api.TempAPI
			case "classic": api.DefaultAPI = &api.ClassicAPI
			case "json": api.DefaultAPI = &api.JsonAPI
			default:
				fmt.Fprintln(os.Stderr, "Invalid API specified.\n" +
					"Valid options are: \"classic\" and \"json\"")
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().BoolVar(&printVersion, "version", false,
		fmt.Sprintf("Print the version (" + Version +") and exit"))
	rootCmd.Flags().StringVarP(&forceAPI, "api", "a", "",
		"Use the specified API for all calls.\n" +
		"Possible options: \"classic\" and \"json\"")

	rootCmd.AddCommand(&cmd.Channel)
	rootCmd.AddCommand(&cmd.Video)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
