package cmd

import (
	"bufio"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
	"os"
	"time"
)

var offset uint

func init() {
	channelDumpCmd.Flags().UintVar(&offset, "page-offset", 1, "Start getting videos at this page. (A page is usually 30 videos)")
}

// The shared context of the request and response threads
var channelDumpContext = struct {
	startTime time.Time
	printResults bool
	writer *bufio.Writer
}{}

// The channel dump route lists
var channelDumpCmd = cobra.Command{
	Use: "dumpurls <channel ID> [file]",
	Short: "Get all public video URLs from channel",
	Long: "Write all videos URLs of a channel to a file",
	Args: cobra.RangeArgs(1, 2),
	Run: cmdFunc(doChannelDump),
}

func doChannelDump(_ *cobra.Command, args []string) error {
	if offset == 0 { offset = 1 }

	printResults := false
	fileName := ""
	channelID := args[0]
	if len(args) != 2 {
		printResults = true
	} else {
		fileName = args[1]
	}
	channelDumpContext.printResults = printResults

	channelID, err := api.GetChannelID(channelID)
	if err != nil { return err }

	log.Infof("Starting work on channel ID \"%s\".", channelID)
	channelDumpContext.startTime = time.Now()

	var flags int
	if force {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	} else {
		flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
	}

	var file *os.File

	if !printResults {
		var err error
		file, err = os.OpenFile(fileName, flags, 0640)
		if err != nil { return err }
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()
		channelDumpContext.writer = writer
	}

	page := offset
	totalURLs := 0

	for {
		// Request next page
		req := apis.Main.GrabChannelPage(channelID, page)

		res := fasthttp.AcquireResponse()
		// TODO defer fasthttp.ReleaseResponse(res)

		err := net.Client.Do(req, res)
		if err != nil {
			log.Errorf("Error at page %d: %v", page, err)
			break
		}

		// Parse response
		channelURLs, err := apis.Main.ParseChannelVideoURLs(res)
		if err != nil { return err }

		// Stop if page is empty
		if len(channelURLs) == 0 { break }

		// Print results
		log.Infof("Received page %d: %d videos.",
			page, len(channelURLs))

		if channelDumpContext.printResults {
			for _, _url := range channelURLs {
				fmt.Println(_url)
			}
		} else {
			for _, _url := range channelURLs {
				_, err := channelDumpContext.writer.WriteString(_url + "\n")
				if err != nil { panic(err) }
			}
		}

		totalURLs += len(channelURLs)
		page++
	}

	duration := time.Since(channelDumpContext.startTime)
	log.Infof("Got %d URLs in %s.", totalURLs, duration.String())
	os.Exit(0)

	return nil
}
