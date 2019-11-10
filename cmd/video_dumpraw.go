package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/apijson"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/valyala/fasthttp"
)

var videoDumpRawCmd = cobra.Command{
	Use: "dumpraw",
	Short: "Get a data set of videos in raw tar format",
	Long: "Reads video IDs from stdin and writes a tar archive with video JSONs to stdout",
	Args: cobra.NoArgs,
	Run: cmdFunc(doVideoRawDump),
}

type videoDumpRaw struct {
	ctx context.Context
	routines sync.WaitGroup
	count int64
}

type rawVideo struct {
	ID  string
	Buf []byte
}

func doVideoRawDump(c *cobra.Command, args []string) (err error) {
	startTime := time.Now()

	videoIDs := make(chan string)
	videos := make(chan rawVideo)

	var d videoDumpRaw
	go func() {
		defer close(videoIDs)
		scn := bufio.NewScanner(os.Stdin)
		for scn.Scan() {
			videoIDs <- scn.Text()
		}
	}()

	// Spawn workers
	d.routines.Add(int(net.MaxWorkers))
	for i := 0; i < int(net.MaxWorkers); i++ {
		go d.videoDumpRawWorker(videos, videoIDs)
	}

	// Print stats
	go func() {
		for range time.Tick(time.Second) {
			logrus.WithField("count", atomic.LoadInt64(&d.count)).Info("Progress")
		}
	}()

	// Tar output
	d.routines.Add(1)
	go func() {
		defer d.routines.Done()
		out := tar.NewWriter(os.Stdout)
		defer func() {
			if err := out.Flush(); err != nil {
				logrus.Error(err)
			}
			if err := out.Close(); err != nil {
				logrus.Error(err)
			}
		}()
		for video := range videos {
			header := tar.Header{
				Name:       video.ID + ".json",
				Size:       int64(len(video.Buf)),
				Mode:       0644,
				ModTime:    time.Now(),
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

	// Print success message
	logrus.Infof("Downloaded %d videos in %s",
		atomic.LoadInt64(&d.count),
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
				Error("Failed to parse video")
			continue
		}
		videos <- rawVideo{
			ID:  videoID,
			Buf: buf,
		}
		atomic.AddInt64(&d.count, 1)
	}
}

func (d *videoDumpRaw) videoDumpSingle(videoID string) ([]byte, error) {
	// Download video info
	req := apis.Main.GrabVideo(videoID)
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
		return nil, apijson.ErrRateLimit
	}

	// Download response
	var buf bytes.Buffer
	err := res.BodyWriteTo(&buf)
	return buf.Bytes(), err
}
