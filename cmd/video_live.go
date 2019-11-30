package cmd

import (
	"encoding/json"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var VideoLive = cobra.Command{
	Use:   "live",
	Short: "Get information about a livestream",
}

func init() {
	VideoLive.AddCommand(
		&videoLiveChatCmd,
	)
}

var videoLiveChatCmd = cobra.Command{
	Use:   "chat",
	Short: "Follow the live chat",
	Args:  cobra.ExactArgs(1),
	Run:   cmdFunc(doVideoLiveChat),
}

func doVideoLiveChat(_ *cobra.Command, args []string) error {
	videoID := args[0]

	videoID, err := api.GetVideoID(videoID)
	if err != nil {
		return err
	}

	startReq := api.GrabLiveChatStart(videoID)
	startRes := fasthttp.AcquireResponse()
	err = net.Client.Do(startReq, startRes)
	if err != nil {
		return err
	}
	fasthttp.ReleaseRequest(startReq)

	enc := json.NewEncoder(os.Stdout)
	enc.SetEscapeHTML(false)

	var out []data.LiveChatMessage
	var cont data.LiveChatContinuation
	cont, err = api.ParseLiveChatStart(&out, startRes)
	if err != nil {
		return err
	}
	fasthttp.ReleaseResponse(startRes)
	for _, msg := range out {
		if err := enc.Encode(&msg); err != nil {
			panic(err)
		}
	}

	for cont.Continuation != "" {
		time.Sleep(time.Duration(cont.Timeout) * time.Millisecond)
		pageReq := api.GrabLiveChatContinuation(cont.Continuation)
		pageRes := fasthttp.AcquireResponse()
		err = net.Client.Do(pageReq, pageRes)
		if err != nil {
			return err
		}
		fasthttp.ReleaseRequest(pageReq)
		cont, err = api.ParseLiveChatPage(&out, pageRes)
		if err != nil {
			return err
		}
		fasthttp.ReleaseResponse(pageRes)
		for _, msg := range out {
			if err := enc.Encode(&msg); err != nil {
				panic(err)
			}
		}
	}

	return nil
}
