package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"fmt"
	"os"
)

var Worker = cobra.Command{
	Use: "worker [config file]",
	Short: "Connect to a queue and start archiving",
	Long: "Get work from a Redis queue, start extracting metadata\n" +
		"and upload it to a Mongo database.",
	Args: cobra.MaximumNArgs(1),
	Run: doWork,
}

func doWork(_ *cobra.Command, args []string) {
	var overrideFile string
	if len(args) == 1 { overrideFile = args[0] }
	err := readConfig(overrideFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func readConfig(overrideFile string) error {
	if overrideFile != "" {
		confFile, err := os.Open(overrideFile)
		if err != nil { return err }
		viper.ReadConfig(confFile)
		return nil
	} else {
		viper.SetConfigName("worker")
		addConfigPaths()
		return viper.ReadInConfig()
	}
}
