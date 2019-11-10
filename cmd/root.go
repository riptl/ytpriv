package cmd

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"os"
)

const Version = "v0.3 -- dev"

var forceAPI string
var concurrentRequests uint
var logLevel string

var Root = cobra.Command{
	Use:   "yt-mango",
	Short: "YT-Mango is a scalable video metadata archiver",
	Long: "YT-Mango is a scalable video metadata archiving utility\n" +
		"written by terorie",
	PersistentPreRun: rootPreRun,
}

func init() {
	Root.Flags().Bool("version", false,
		fmt.Sprintf("Print the version (" + Version +") and exit"))

	pf := Root.PersistentFlags()
	pf.StringVarP(&forceAPI, "api", "a", "",
		"Use the specified API for all calls.\n" +
		"Possible options: \"classic\" and \"json\"")
	pf.UintVarP(&concurrentRequests, "concurrency", "c", 4,
		"Number of maximum concurrent HTTP requests")
	pf.StringVarP(&logLevel, "log-level", "l", "",
		"Log level. Valid options are:\n" +
		"{debug, info, warn, error, fatal, panic}")
	pf.StringVar(&net.Client.Name, "user-agent", "yt-mango/0.1",
		"HTTP client user-agent")

	Root.AddCommand(&Channel)
	Root.AddCommand(&Video)
}

func rootPreRun(_ *cobra.Command, _ []string) {
	net.MaxWorkers = concurrentRequests
	net.Client.MaxConnsPerHost = int(concurrentRequests)

	switch forceAPI {
	case "": apis.Main = &apis.TempAPI
	case "classic": apis.Main = &apis.ClassicAPI
	case "json": apis.Main = &apis.JsonAPI
	default:
		fmt.Fprintln(os.Stderr, "Invalid API specified.\n"+
			"Valid options are: \"classic\" and \"json\"")
		os.Exit(1)
	}

	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	}
}
