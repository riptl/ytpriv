package cmd

import (
	"github.com/spf13/cobra"
	"net/url"
	"fmt"
	"os"
	"strings"
	"time"
	"bufio"
	"log"
	"github.com/terorie/yt-mango/api"
)

var channelDumpCmd = cobra.Command{
	Use: "dumpurls <channel ID> <file>",
	Short: "Get all public video URLs from channel",
	Long: "Write all videos URLs of a channel to a file",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		channelID := args[0]
		fileName := args[1]

		if !matchChannelID.MatchString(channelID) {
			// Check if youtube.com domain
			_url, err := url.Parse(channelID)
			if err != nil || (_url.Host != "www.youtube.com" && _url.Host != "youtube.com") {
				fmt.Fprintln(os.Stderr, "Not a channel ID:", channelID)
				os.Exit(1)
			}

			// Check if old /user/ URL
			if strings.HasPrefix(_url.Path, "/user/") {
				// TODO Implement extraction of channel ID
				fmt.Fprintln(os.Stderr, "New /channel/ link is required!\n" +
					"The old /user/ links do not work.")
				os.Exit(1)
			}

			// Remove /channel/ path
			channelID = strings.TrimPrefix(_url.Path, "/channel/")
			if len(channelID) == len(_url.Path) {
				// No such prefix to be removed
				fmt.Fprintln(os.Stderr, "Not a channel ID:", channelID)
				os.Exit(1)
			}

			// Remove rest of path from channel ID
			slashIndex := strings.IndexRune(channelID, '/')
			if slashIndex != -1 {
				channelID = channelID[:slashIndex]
			}
		}

		log.Printf("Starting work on channel ID \"%s\".", channelID)
		startTime := time.Now()

		var flags int
		if force {
			flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		} else {
			flags = os.O_WRONLY | os.O_CREATE | os.O_EXCL
		}

		file, err := os.OpenFile(fileName, flags, 0640)
		defer file.Close()
		writer := bufio.NewWriter(file)
		defer writer.Flush()

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		totalURLs := 0
		for i := offset; true; i++ {
			channelURLs, err := api.DefaultAPI.GetChannelVideoURLs(channelID, uint(i))
			if err != nil {
				log.Printf("Aborting on error %v.", err)
				break
			}
			if len(channelURLs) == 0 {
				log.Printf("Page %d returned no videos.", i)
				break
			}
			totalURLs += len(channelURLs)
			log.Printf("Received page %d: %d videos.", i, len(channelURLs))

			for _, _url:= range channelURLs {
				_, err := writer.WriteString(_url + "\n")
				if err != nil { panic(err) }
			}
		}

		duration := time.Since(startTime)
		log.Printf("Got %d URLs in %s.", totalURLs, duration.String())
	},
}
