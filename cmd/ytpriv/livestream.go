package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/ytpriv"
	"github.com/terorie/ytpriv/types"
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
	flags := livestreamChat.Flags()
	flags.Bool("top", false, "Top chat only")
}

func doLivestreamChat(c *cobra.Command, args []string) error {
	flags := c.Flags()
	top, err := flags.GetBool("top")
	if err != nil {
		panic(err.Error())
	}
	videoID := args[0]
	videoID, err = yt.ExtractVideoID(videoID)
	if err != nil {
		return err
	}
	video, err := client.RequestVideo(videoID).Do()
	if err != nil {
		return fmt.Errorf("failed to get video info: %w", err)
	}
	var cont string
	var live bool
	if top {
		live = video.TopChatContinuation != ""
		if live {
			cont = video.TopChatContinuation
		} else {
			cont = video.TopChatReplayContinuation
		}
	} else {
		live = video.LiveChatContinuation != ""
		if live {
			cont = video.LiveChatContinuation
		} else {
			cont = video.LiveChatReplayContinuation
		}
	}
	if cont == "" {
		return fmt.Errorf("no live chat found")
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)
	if !live {
		logrus.Info("Dumping live chat replay")
		for i := 0; cont != ""; i++ {
			logrus.WithField("page", i).Info("Dumping page")
			var out []*types.LivechatMessage
			out, cont, err = client.RequestLivechatReplay(cont).Do()
			if err != nil {
				return err
			}
			for _, msg := range out {
				if err := enc.Encode(&msg); err != nil {
					panic(err)
				}
			}
		}
	} else {
		logrus.Info("Dumping live chat")
		liveCont := yt.LivechatContinuation{
			Timeout:      0,
			Continuation: cont,
		}
		for i := 0; liveCont.Continuation != ""; i++ {
			logrus.
				WithField("page", i).
				WithField("timeout_ms", liveCont.Timeout).
				Info("Dumping page")
			time.Sleep(time.Duration(liveCont.Timeout) * time.Millisecond)
			var out []*types.LivechatMessage
			out, liveCont, err = client.RequestLivechat(liveCont.Continuation).Do()
			if err != nil {
				return err
			}
			for _, msg := range out {
				if err := enc.Encode(&msg); err != nil {
					panic(err)
				}
			}
		}
	}
	logrus.Info("Reached end of chat")
	return nil
}
