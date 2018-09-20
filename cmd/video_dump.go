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
	Args: cobra.ExactArgs(3),
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
	nStr := args[0]
	path := args[1]
	firstID := args[2]

	toFindUint64, err := strconv.ParseUint(nStr, 10, 31)
	if err != nil { return }
	toFind := int32(toFindUint64)
	left := toFind

	err = os.Mkdir(path, 0755)
	if err != nil { return }

	firstID, err = api.GetVideoID(firstID)
	if err != nil { return }

	d := videoDump{
		path:     path,
		left:     left,
		queue:    []string{firstID},
		foundAll: 0,
	}

	// Insert first video ID into buffer
	d.queueLock.Lock()
	d.queue = []string{firstID}
	d.queueLock.Unlock()

	// Spawn workers
	for i := 0; i < int(net.MaxWorkers); i++ {
		d.activeRoutines.Add(1)
		go d.videoDumpWorker()
	}

	// Wait for routines to finish
	d.activeRoutines.Wait()

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
		videoId, ok := d.getNextID()
		if !ok {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if d.videoDumpSingle(videoId) {
			// Update videos left counter
			if atomic.AddInt32(&d.left, -1) <= 0 {
				// No vids left
				println("waiting")
				return
			}
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

	logrus.WithFields(logrus.Fields{
		"id": videoId,
		"path": vidPath,
	}).Info("Got video")

	success = true
	return
}
