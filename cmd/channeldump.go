package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"time"
	"bufio"
	"log"
	"github.com/terorie/yt-mango/api"
	"fmt"
	"github.com/terorie/yt-mango/common"
	"sync/atomic"
	"errors"
)

var channelDumpContext = struct{
	startTime time.Time
	printResults bool
	writer *bufio.Writer
	pagesDone uint64
	errorOccured int32 // Use atomic boolean here
}{}

var channelDumpCmd = cobra.Command{
	Use: "dumpurls <channel ID> [file]",
	Short: "Get all public video URLs from channel",
	Long: "Write all videos URLs of a channel to a file",
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
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
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}

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

		results := make(chan common.JobResult)
		terminateSub := make(chan bool)

		// TODO Clean up
		go processResults(results, terminateSub)

		page := offset
		for {
			// Terminate if error detected
			if atomic.LoadInt32(&channelDumpContext.errorOccured) != 0 {
				goto terminate
			}
			// Send new requests
			req := api.Main.GrabChannelPage(channelID, page)
			common.DoAsyncHTTP(req, results, page)

			page++
		}
		terminate:

		// Requests sent, wait for remaining requests to finish
		for {
			done := uint64(offset) + atomic.LoadUint64(&channelDumpContext.pagesDone)
			target := uint64(page)
			if done >= target { break }

			// TODO use semaphore
			time.Sleep(time.Millisecond)
		}

		// TODO Don't ignore pending results
		duration := time.Since(channelDumpContext.startTime)
		log.Printf("Done in %s.", duration.String())

		terminateSub <- true
	},
}

// TODO combine channels into one
func processResults(results chan common.JobResult, terminateSub chan bool) {
	totalURLs := 0
	for {
		select {
		case <-terminateSub:
			log.Printf("Got %d URLs", totalURLs)
			os.Exit(0)
			return
		case res := <-results:
			var err error
			var channelURLs []string
			page := res.ReqData.(uint)
			if res.Err != nil {
				err = res.Err
				goto endError
			}
			channelURLs, err = api.Main.ParseChannelVideoURLs(res.Res)
			if err != nil { goto endError }
			if len(channelURLs) == 0 {
				err = errors.New("returned no videos")
				goto endError
			}
			totalURLs += len(channelURLs)
			log.Printf("Received page %d: %d videos.", page, len(channelURLs))

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
			// Increment done pages count
			atomic.AddUint64(&channelDumpContext.pagesDone, 1)
			continue
			endError:
				atomic.AddUint64(&channelDumpContext.pagesDone, 1)
				atomic.StoreInt32(&channelDumpContext.errorOccured, 1)
				log.Printf("Error at page %d: %v", page, err)
		}
	}
}
