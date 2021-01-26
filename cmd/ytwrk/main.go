package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yt "github.com/terorie/ytpriv"
	"github.com/valyala/fasthttp"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

const version = "v1.0.0"

var concurrentRequests uint
var logLevel string

var rootCmd = cobra.Command{
	Use:              "ytpriv",
	Short:            "ytpriv is a YouTube metadata exporter",
	Long:             "https://github.com/terorie/ytpriv",
	PersistentPreRun: rootPreRun,
}

func init() {
	pf := rootCmd.PersistentFlags()
	pf.UintVarP(&concurrentRequests, "concurrency", "c", 4,
		"Number of maximum concurrent HTTP requests")
	pf.StringVarP(&logLevel, "log-level", "l", "",
		"Log level. Valid options are:\n"+
			"{debug, info, warn, error, fatal, panic}")
	pf.StringVar(&client.HTTP.Name, "user-agent", "ytpriv/"+version,
		"HTTP client user-agent")
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
		if _, err := os.Stat("/tmp/ytpriv_high_rate"); os.IsNotExist(err) {
			logrus.Warn("If you know what you are doing, touch /tmp/ytpriv_high_rate")
			logrus.Fatal("Terminating.")
		}
	}

	maxWorkers = concurrentRequests
	client.HTTP.MaxConnsPerHost = int(concurrentRequests)

	if logLevel != "" {
		lvl, err := logrus.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		logrus.SetLevel(lvl)
	}
}

var maxWorkers uint

var client = yt.Client{
	HTTP: &fasthttp.Client{
		Name:                          "ytpriv/v0.4",
		DisableHeaderNamesNormalizing: true,
		MaxConnsPerHost:               50,
	},
}
