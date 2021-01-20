package main

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
	yt "github.com/terorie/ytwrk"
	"github.com/valyala/fasthttp"
)

var channelDumpRawCmd = cobra.Command{
	Use:   "dumpraw",
	Short: "Get a data set of channels in raw tar format",
	Long:  "Reads channel IDs from stdin and writes a tar archive with channel JSONs to stdout",
	Args:  cobra.NoArgs,
	Run:   cmdFunc(doChannelRawDump),
}

type channelDumpRaw struct {
	routines sync.WaitGroup
	count    expvar.Int
}

type rawChannel struct {
	ID  string
	Buf []byte
}

func doChannelRawDump(c *cobra.Command, _ []string) (err error) {
	startTime := time.Now()

	channelIDs := make(chan string)
	channels := make(chan rawChannel)

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
		defer close(channelIDs)
		scn := bufio.NewScanner(os.Stdin)
		for atomic.LoadUint32(&externalStop) == 0 && scn.Scan() {
			channelIDs <- scn.Text()
		}
	}()

	// Spawn workers
	var d channelDumpRaw
	d.routines.Add(int(maxWorkers))
	for i := 0; i < int(maxWorkers); i++ {
		go d.channelDumpRawWorker(channels, channelIDs)
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
		for channel := range channels {
			header := tar.Header{
				Name:    channel.ID + ".json",
				Size:    int64(len(channel.Buf)),
				Mode:    0644,
				ModTime: time.Now(),
			}
			if err := out.WriteHeader(&header); err != nil {
				logrus.Fatal(err)
			}
			if _, err := out.Write(channel.Buf); err != nil {
				logrus.Fatal(err)
			}
		}
	}()

	// Wait for routines to finish
	d.routines.Wait()
	close(channels)
	printer.Wait()

	// Print success message
	logrus.Infof("Downloaded %d channels in %s",
		d.count.Value(),
		time.Since(startTime).String())

	return nil
}

func (d *channelDumpRaw) channelDumpRawWorker(channels chan<- rawChannel, channelIDs <-chan string) {
	defer d.routines.Done()
	for channelID := range channelIDs {
		buf, err := d.channelDumpSingle(channelID)
		if err != nil {
			logrus.WithError(err).
				WithField("id", channelID).
				Error("Failed to get channel")
			continue
		}
		channels <- rawChannel{
			ID:  channelID,
			Buf: buf,
		}
		d.count.Add(1)
	}
}

func (d *channelDumpRaw) channelDumpSingle(channelID string) ([]byte, error) {
	// Start request
	req := client.RequestChannel(channelID)
	defer fasthttp.ReleaseRequest(req)
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)
	if err := client.HTTP.Do(req, res); err != nil {
		logrus.WithError(err).
			WithField("id", channelID).
			Error("Failed to download channel")
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
		return nil, yt.ErrRateLimit
	}
	// Download response
	var buf bytes.Buffer
	err := res.BodyWriteTo(&buf)
	return buf.Bytes(), err
}
