package cmd

import (
	"fmt"
	"os"
	"bufio"
	"github.com/spf13/cobra"
		"github.com/terorie/yt-mango/net"
	"github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/apis"
)

const Version = "v0.1 -- dev"

var forceAPI string
var concurrentRequests uint
var debugHttpFile string
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
	pf.StringVar(&debugHttpFile, "debug-file", "",
		"Log all HTTP actions to a JSON-like file\n" +
		"(one request/response pair per line)")
	pf.StringVarP(&logLevel, "log-level", "l", "",
		"Log level. Valid options are:\n" +
		"{debug, info, warn, error, fatal, panic}")

	Root.AddCommand(&Channel)
	Root.AddCommand(&Video)
	Root.AddCommand(&DebugFile)
	Root.AddCommand(&Worker)
}

func rootPreRun(_ *cobra.Command, _ []string) {
	net.MaxWorkers = uint32(concurrentRequests)

	if debugHttpFile != "" {
		debugFile, err := os.OpenFile(debugHttpFile,
			os.O_WRONLY | os.O_CREATE | os.O_APPEND,
			0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Could not open HTTP debug file:", err)
			os.Exit(1)
		}

		debugWriter := bufio.NewWriter(debugFile)

		// Force all HTTP requests through debug code
		net.Client.Transport = net.DebugTransport{ debugFile, debugWriter }
	}

	switch forceAPI {
	case "": apis.Main = &apis.TempAPI
	case "classic": apis.Main = &apis.ClassicAPI
	case "json": apis.Main = &apis.JsonAPI
	default:
		fmt.Fprintln(os.Stderr, "Invalid API specified.\n" +
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
