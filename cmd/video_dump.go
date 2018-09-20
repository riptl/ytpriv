package cmd

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/terorie/yt-mango/api"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var videoDumpCmd = cobra.Command{
	Use: "dump <n> <path> <video>",
	Short: "Get a data set of videos",
	Long: "Loads n video metadata files to the\n" +
		"specified path starting at the root video.\n" +
		"The videos are saved as \"<video_id>.json\".",
	Args: cobra.MinimumNArgs(3),
	Run: cmdFunc(doVideoDump),
}

type videoDump struct{
	path           string
	left           int32
	found          sync.Map
	queueLock      sync.Mutex
	queue          []string
	foundAll       int32
	activeRoutines sync.WaitGroup
}

func doVideoDump(_ *cobra.Command, args []string) (err error) {
	startTime := time.Now()

	nStr := args[0]
	path := args[1]
	startIDs := args[2:]

	toDownload, err := strconv.ParseUint(nStr, 10, 31)
	if err != nil { return }

	err = os.Mkdir(path, 0755)
	if err != nil { return }

	for i, url := range startIDs {
		startIDs[i], err = api.GetVideoID(url)
		if err != nil { return }
	}

	d := videoDump{
		path:     path,
		left:     int32(toDownload),
		queue:    startIDs,
		foundAll: 0,
	}

	// Spawn workers
	for i := 0; i < int(net.MaxWorkers); i++ {
		d.activeRoutines.Add(1)
		go d.videoDumpWorker()
	}

	// Print stats
	go func() {
		for range time.NewTicker(time.Second).C {
			_left := atomic.LoadInt32(&d.left)
			if _left != 0 {
				logrus.WithField("left", _left).
					Infof("%d videos left", _left)
			} else {
				break
			}
		}
	}()

	// Wait for routines to finish
	d.activeRoutines.Wait()

	// Print success message
	logrus.Infof("Downloaded %d videos in %s",
		toDownload,
		time.Since(startTime).String())

	return nil
}

func (d *videoDump) getNextID() (s string, ok bool) {
	d.queueLock.Lock()
	defer d.queueLock.Unlock()

	if len(d.queue) > 0 {
		ok = true
		s = d.queue[0]
		d.queue = d.queue[1:]
	} else {
		ok = false
	}
	return
}

func (d *videoDump) videoDumpWorker() {
	defer d.activeRoutines.Done()

	for {
		// Check if all videos loaded
		if atomic.LoadInt32(&d.left) == 0 { return }

		videoId, ok := d.getNextID()
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// Update videos left counter
		if atomic.AddInt32(&d.left, -1) <= 0 { return }

		if !d.videoDumpSingle(videoId) {
			// Increment videos left counter back again
			atomic.AddInt32(&d.left, +1)
		}
	}
}

func (d *videoDump) videoDumpSingle(videoId string) (success bool) {
	success = false

	// Download video info
	req := apis.Main.GrabVideo(videoId)
	res, err := net.Client.Do(req)
	if err != nil {
		logrus.WithError(err).
			WithField("id", videoId).
			Error("Failed to download video")
		return
	}

	var v data.Video
	v.ID = videoId

	// Parse video
	err = apis.Main.ParseVideo(&v, res)
	if err != nil {
		logrus.WithError(err).
			WithField("id", videoId).
			Error("Failed to parse video")
		return
	}

	var newIDs []string
	for _, id := range v.Related {
		if _, loaded := d.found.LoadOrStore(id, true); !loaded {
			newIDs = append(newIDs, id)
		}
	}

	// Add newly found video IDs
	if atomic.LoadInt32(&d.foundAll) == 0 {
		d.queueLock.Lock()
		d.queue = append(d.queue, newIDs...)
		d.queueLock.Unlock()
	}

	// Open file
	vidPath := filepath.Join(d.path, videoId + ".json")
	file, err := os.OpenFile(
		vidPath,
		os.O_CREATE | os.O_EXCL | os.O_WRONLY,
		0644,
	)
	if err != nil {
		logrus.WithError(err).
			WithField("path", vidPath).
			Error("Failed to create file")
	}
	defer file.Close()

	// Save video to JSON
	enc := json.NewEncoder(file)
	enc.SetIndent("", "\t")
	err = enc.Encode(&v)
	if err != nil { return }

	logrus.WithField("id", videoId).Debug("Got video")

	success = true
	return
}
