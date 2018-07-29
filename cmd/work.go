package cmd

import "github.com/spf13/cobra"

var Work = cobra.Command{
	Use: "work",
	Short: "Connect to a queue and start archiving",
	Long: "Get work from a Redis queue, start extracting metadata\n" +
		"and upload it to a Mongo database.",
}
