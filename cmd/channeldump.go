package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"time"
	"bufio"
	"log"
	"github.com/terorie/yt-mango/api"
	"fmt"
	"github.com/terorie/yt-mango/net"
	"sync/atomic"
	"errors"
	"sync"
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
	// Number of pages that have been
	// requested but not yet received.
	// Additional +1 is added if additional
	// are planned to be requested
	pagesToReceive sync.WaitGroup
	// If set to non-zero, an error was received
	errorOccurred int32
}{}

// The channel dump route lists
var channelDumpCmd = cobra.Command{
	Use: "dumpurls <channel ID> [file]",
	Short: "Get all public video URLs from channel",
	Long: "Write all videos URLs of a channel to a file",
	Args: cobra.RangeArgs(1, 2),
	Run: doChannelDump,
}

func doChannelDump(_ *cobra.Command, args []string) {
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

	channelID = api.GetChannelID(channelID)
	if channelID == "" { os.Exit(1) }

	log.Printf("Starting work on channel ID \"%s\".", channelID)
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
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		defer writer.Flush()
		channelDumpContext.writer = writer
	}

	results := make(chan net.JobResult)
	terminateSub := make(chan bool)

	// TODO Clean up
	go channelDumpResults(results, terminateSub)

	page := offset
	for {
		// Terminate if error detected
		if atomic.LoadInt32(&channelDumpContext.errorOccurred) != 0 {
			goto terminate
		}
		// Send new requests
		req := api.Main.GrabChannelPage(channelID, page)
		channelDumpContext.pagesToReceive.Add(1)
		net.DoAsyncHTTP(req, results, page)

		page++
	}
	terminate:

	// Requests sent, wait for remaining requests to finish
	channelDumpContext.pagesToReceive.Wait()

	terminateSub <- true
}

	// Helper goroutine that processes HTTP results.
// HTTP results are received on "results".
// The routine exits if a value on "terminateSub" is received.
// For every incoming result (error or response),
// the "pagesToReceive" counter is decreased.
// If an error is received, the "errorOccurred" flag is set.
func channelDumpResults(results chan net.JobResult, terminateSub chan bool) {
	totalURLs := 0
	for {
		select {
		case <-terminateSub:
			duration := time.Since(channelDumpContext.startTime)
			log.Printf("Got %d URLs in %s.", totalURLs, duration.String())
			os.Exit(0)
			return
		case res := <-results:
			page, numURLs, err := channelDumpResult(&res)
			// Mark page as processed
			channelDumpContext.pagesToReceive.Done()
			// Report back error
			if err != nil {
				atomic.StoreInt32(&channelDumpContext.errorOccurred, 1)
				log.Printf("Error at page %d: %v", page, err)
			} else {
				totalURLs += numURLs
			}
		}
	}
}

// Processes a HTTP result
func channelDumpResult(res *net.JobResult) (page uint, numURLs int, err error) {
	var channelURLs []string

	// Extra data is page number
	page = res.ReqData.(uint)
	// Abort if request failed
	if res.Err != nil { return page, 0, res.Err }

	// Parse response
	channelURLs, err = api.Main.ParseChannelVideoURLs(res.Res)
	if err != nil { return }
	numURLs = len(channelURLs)
	if numURLs == 0 { return page, 0, errors.New("returned no videos") }

	// Print results
	log.Printf("Received page %d: %d videos.", page, numURLs)

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

	return
}
