package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"expvar"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/ytwrk/api"
	"github.com/terorie/ytwrk/net"
	"github.com/valyala/fasthttp"
)

var videoDumpRawCmd = cobra.Command{
	Use:   "dumpraw",
	Short: "Get a data set of videos in raw tar format",
	Long:  "Reads video IDs from stdin and writes a tar archive with video JSONs to stdout",
	Args:  cobra.NoArgs,
	Run:   cmdFunc(doVideoRawDump),
}

type videoDumpRaw struct {
	routines sync.WaitGroup
	count    expvar.Int
}

type rawVideo struct {
	ID  string
	Buf []byte
}

func doVideoRawDump(c *cobra.Command, _ []string) (err error) {
	startTime := time.Now()

	videoIDs := make(chan string)
	videos := make(chan rawVideo)

	// Listen for Ctrl+C
	var externalStop uint32
	go func() {
		ctx := signalContext(context.Background())
		<-ctx.Done()
		logrus.Info("Shutting down")
		atomic.StoreUint32(&externalStop, 1)
		_ = os.Stdin.Close()
	}()

	// Read input
	go func() {
		defer close(videoIDs)
		scn := bufio.NewScanner(os.Stdin)
		for atomic.LoadUint32(&externalStop) == 0 && scn.Scan() {
			videoIDs <- scn.Text()
		}
	}()

	// Spawn workers
	var d videoDumpRaw
	d.routines.Add(int(net.MaxWorkers))
	for i := 0; i < int(net.MaxWorkers); i++ {
		go d.videoDumpRawWorker(videos, videoIDs)
	}

	// Print stats
	go func() {
		for range time.Tick(time.Second) {
			logrus.WithField("count", d.count.Value()).Info("Progress")
		}
	}()

	// Tar output
	var printer sync.WaitGroup
	printer.Add(1)
	go func() {
		defer printer.Done()
		out := tar.NewWriter(os.Stdout)
		defer func() {
			if err := out.Flush(); err != nil {
				logrus.Error(err)
			}
			if err := out.Close(); err != nil {
				logrus.Error(err)
			}
			logrus.Info("Closed tar stream")
		}()
		for video := range videos {
			header := tar.Header{
				Name:    video.ID + ".json",
				Size:    int64(len(video.Buf)),
				Mode:    0644,
				ModTime: time.Now(),
			}
			if err := out.WriteHeader(&header); err != nil {
				logrus.Fatal(err)
			}
			if _, err := out.Write(video.Buf); err != nil {
				logrus.Fatal(err)
			}
		}
	}()

	// Wait for routines to finish
	d.routines.Wait()
	close(videos)
	printer.Wait()

	// Print success message
	logrus.Infof("Downloaded %d videos in %s",
		d.count.Value(),
		time.Since(startTime).String())

	return nil
}

func (d *videoDumpRaw) videoDumpRawWorker(videos chan<- rawVideo, videoIDs <-chan string) {
	defer d.routines.Done()
	for videoID := range videoIDs {
		buf, err := d.videoDumpSingle(videoID)
		if err != nil {
			logrus.WithError(err).
				WithField("id", videoID).
				Error("Failed to get video")
			continue
		}
		videos <- rawVideo{
			ID:  videoID,
			Buf: buf,
		}
		d.count.Add(1)
	}
}

func (d *videoDumpRaw) videoDumpSingle(videoID string) ([]byte, error) {
	// Download video info
	req := api.GrabVideo(videoID)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	// Start request
	if err := net.Client.Do(req, res); err != nil {
		logrus.WithError(err).
			WithField("id", videoID).
			Error("Failed to download video")
		return nil, err
	}

	// Check response headers
	if res.StatusCode() != fasthttp.StatusOK {
		return nil, fmt.Errorf("response status %d", res.StatusCode())
	}
	contentType := res.Header.ContentType()
	switch {
	case bytes.HasPrefix(contentType, []byte("application/json")):
		break
	case bytes.HasPrefix(contentType, []byte("text/html")):
		return nil, api.ErrRateLimit
	}

	// Download response
	var buf bytes.Buffer
	err := res.BodyWriteTo(&buf)
	return buf.Bytes(), err
}
