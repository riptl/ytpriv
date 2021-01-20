package main

import (
	"bufio"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	yt "github.com/terorie/ytwrk"
	"github.com/valyala/fasthttp"
)

var channelDumpCmd = cobra.Command{
	Use:   "dump [channel ID...]",
	Short: "Get all public video URLs from channels",
	Long: "Takes in channel IDs and produces a list of video IDs uploaded by them.\n" +
		"If no channel IDs are given as arguments, they are read from stdin.",
	Args: cobra.ArbitraryArgs,
	Run:  cmdFunc(doChannelDump),
}

func doChannelDump(_ *cobra.Command, args []string) error {
	start := time.Now()

	// Create argument channels
	jobs := make(chan string)
	go stdinOrArgs(jobs, args)
	channelIDs := make(chan string)
	go func() {
		defer close(channelIDs)
		for job := range jobs {
			channelID, err := yt.ExtractChannelID(job)
			if err != nil {
				log.Error(err)
				continue
			}
			channelIDs <- channelID
		}
	}()

	// Create videoIDs from channel IDs
	wr := bufio.NewWriter(os.Stdout)
	results := make(chan [2]string)
	go channelDumpScheduler(results, channelIDs)
	for result := range results {
		// Avoid fmt newline flushes
		_, _ = wr.WriteString(result[0])
		_ = wr.WriteByte('\t')
		_, _ = wr.WriteString(result[1])
		_ = wr.WriteByte('\n')
	}
	_ = wr.Flush()

	log.Infof("Finished after %s", time.Since(start).String())

	return nil
}

func channelDumpScheduler(results chan<- [2]string, channelIDs <-chan string) {
	var wg sync.WaitGroup
	wg.Add(int(maxWorkers))
	for i := uint(0); i < maxWorkers; i++ {
		go func() {
			var errors int
			defer wg.Done()
			for channelID := range channelIDs {
				log.Infof("Dumping channel %s", channelID)
				err := dumpChannel(results, channelID)
				if err != nil {
					errors++
					log.WithError(err).Errorf("Failed to dump channel %s", channelID)
				} else {
					errors = 0
				}
				time.Sleep(time.Second * time.Duration(errors))
			}
		}()
	}
	wg.Wait()
	close(results)
}

func dumpChannel(results chan<- [2]string, channelID string) error {
	channelID, err := yt.ExtractChannelID(channelID)
	if err != nil {
		return err
	}

	page := uint(0)
	totalURLs := 0

	for {
		// Request next page
		req := client.RequestChannelPage(channelID, page)
		res := fasthttp.AcquireResponse()
		err := client.HTTP.Do(req, res)
		fasthttp.ReleaseRequest(req)
		if err != nil {
			log.Errorf("Error at page %d: %v", page, err)
			break
		}
		// Parse response
		videoURLs, err := yt.ParseChannelVideoURLs(res)
		fasthttp.ReleaseResponse(res)
		if err != nil {
			return err
		}
		// Stop if page is empty
		if len(videoURLs) == 0 {
			break
		}
		// Print results
		log.WithFields(log.Fields{
			"channel": channelID,
			"page":    page,
			"videos":  len(videoURLs),
		}).Info("Got page")
		for _, _url := range videoURLs {
			videoID, err := yt.ExtractVideoID(_url)
			if err != nil {
				log.WithError(err).Warn("Got invalid video URL")
			}
			results <- [2]string{channelID, videoID}
		}
		totalURLs += len(videoURLs)
		page++
	}
	return nil
}
