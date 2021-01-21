package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk"
)

var livestreamCmd = cobra.Command{
	Use:   "livestream",
	Short: "Scrape a livestream",
}

func init() {
	rootCmd.AddCommand(&livestreamCmd)
}

var livestreamChat = cobra.Command{
	Use:   "chat",
	Short: "Follow the live chat",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doLivestreamChat),
}

func init() {
	livestreamCmd.AddCommand(&livestreamChat)
}

func doLivestreamChat(_ *cobra.Command, args []string) error {
	videoID := args[0]
	videoID, err := yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}

	out, cont, err := client.RequestLivechatStart(videoID).Do()
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	for _, msg := range out {
		if err := enc.Encode(&msg); err != nil {
			panic(err)
		}
	}

	for cont.Continuation != "" {
		time.Sleep(time.Duration(cont.Timeout) * time.Millisecond)
		out, cont, err = client.RequestLivechatContinuation(cont.Continuation).Do()
		if err != nil {
			return err
		}
		for _, msg := range out {
			if err := enc.Encode(&msg); err != nil {
				panic(err)
			}
		}
	}
	return nil
}
