package cmd

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
	"os"
	"sync"
	"time"
)

// The channel dump route lists
var channelDumpCmd = cobra.Command{
	Use:   "dumpurls [channel ID...]",
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
			channelID, err := api.GetChannelID(job)
			if err != nil {
				log.Error(err)
				continue
			}
			channelIDs <- channelID
		}
	}()

	// Create videoIDs from channel IDs
	wr := bufio.NewWriter(os.Stdout)
	videoIDs := make(chan string)
	go channelDumpScheduler(videoIDs, channelIDs)
	for videoID := range videoIDs {
		// Avoid fmt newline flushes
		_, _ = wr.WriteString(videoID)
		_ = wr.WriteByte('\n')
	}

	log.Infof("Finished after %s", time.Since(start).String())

	return nil
}

func channelDumpScheduler(videoIDs chan<- string, channelIDs <-chan string) {
	var wg sync.WaitGroup
	wg.Add(int(net.MaxWorkers))
	for i := uint(0); i < net.MaxWorkers; i++ {
		go func() {
			defer wg.Done()
			for channelID := range channelIDs {
				log.Infof("Dumping channel %s", channelID)
				err := dumpChannel(channelID, videoIDs)
				if err != nil {
					log.WithError(err).Errorf("Failed to dump channel %s", channelID)
				}
			}
		}()
	}
	wg.Wait()
	close(videoIDs)
}

func dumpChannel(channelID string, videoIDs chan<- string) error {
	channelID, err := api.GetChannelID(channelID)
	if err != nil { return err }

	page := uint(0)
	totalURLs := 0

	for {
		// Request next page
		req := apis.Main.GrabChannelPage(channelID, page)
		res := fasthttp.AcquireResponse()

		err := net.Client.Do(req, res)
		fasthttp.ReleaseRequest(req)
		if err != nil {
			log.Errorf("Error at page %d: %v", page, err)
			break
		}

		// Parse response
		videoURLs, err := apis.Main.ParseChannelVideoURLs(res)
		fasthttp.ReleaseResponse(res)
		if err != nil { return err }

		// Stop if page is empty
		if len(videoURLs) == 0 { break }

		// Print results
		log.WithFields(log.Fields{
			"channel": channelID,
			"page":    page,
			"videos":  len(videoURLs),
		}).Info("Got page")

		for _, _url := range videoURLs {
			videoID, err := api.GetVideoID(_url)
			if err != nil {
				log.WithError(err).Warn("Got invalid video URL")
			}
			videoIDs <- videoID
		}

		totalURLs += len(videoURLs)
		page++
	}
	return nil
}
