package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk/net"
)

const Version = "v0.4"

var concurrentRequests uint
var logLevel string

var Root = cobra.Command{
	Use:              "ytwrk",
	Short:            "ytwrk is a fast and light YouTube metadata exporter",
	Long:             "https://github.com/terorie/ytwrk",
	PersistentPreRun: rootPreRun,
}

func init() {
	Root.Flags().Bool("version", false,
		fmt.Sprintf("Print the version ("+Version+") and exit"))

	pf := Root.PersistentFlags()
	pf.UintVarP(&concurrentRequests, "concurrency", "c", 4,
		"Number of maximum concurrent HTTP requests")
	pf.StringVarP(&logLevel, "log-level", "l", "",
		"Log level. Valid options are:\n"+
			"{debug, info, warn, error, fatal, panic}")
	pf.StringVar(&net.Client.Name, "user-agent", "ytwrk/"+Version,
		"HTTP client user-agent")

	Root.AddCommand(
		&Channel,
		&Video,
		&Playlist,
	)
}

func rootPreRun(_ *cobra.Command, _ []string) {
	if concurrentRequests > 32 {
		logrus.Warn("#################")
		logrus.Warn("#### WARNING ####")
		logrus.Warn("#################")
		logrus.Warn("It looks like you are trying to crawl YouTube with a high request rate.")
		logrus.Warn("This is highly discouraged and will likely result in automated permanent IP bans.")
		logrus.Warn("Abusing any service with high request rates forces the operators to implement rate limits")
		logrus.Warn("and fingerprinting bans, hurting tools like this and everyone else relying on those services.")
		if _, err := os.Stat("/tmp/ytwrk_high_rate"); os.IsNotExist(err) {
			logrus.Warn("If you know what you are doing, touch /tmp/ytwrk_high_rate")
			logrus.Fatal("Terminating.")
		}
	}

	net.MaxWorkers = concurrentRequests
	net.Client.MaxConnsPerHost = int(concurrentRequests)

	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	}
}
