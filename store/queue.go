package store

import (
	log "github.com/sirupsen/logrus"
	"time"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/terorie/yt-mango/viperstruct"
	"errors"
)

// Sorted set with VID as key and
// last crawl time (Unix) as score
const videoSet = "VIDEO_SET"

// List with VIDs that are scheduled
// to be crawled
const videoWaitQueue = "VIDEO_WAIT"

var queue *redis.Client

// Redis queue

func ConnectQueue() error {
	// Default config vars
	viper.SetDefault("queue.network", "tcp")
	viper.SetDefault("queue.host", "localhost:6379")
	viper.SetDefault("queue.db", 0)

	var queueConf struct{
		Network string `viper:"queue.network"`
		Host string `viper:"queue.host"`
		Pass string `viper:"queue.pass,optional"`
		DB int `viper:"queue.db"`
	}

	if err := viperstruct.ReadConfig(&queueConf);
		err != nil { return err }

	queue = redis.NewClient(&redis.Options{
		Network: queueConf.Network,
		Addr: queueConf.Host,
		Password: queueConf.Pass,
		DB: queueConf.DB,
	})
	if queue == nil { return errors.New("could not connect to Redis") }

	return nil
}

func DisconnectQueue() {
	if err := queue.Close(); err != nil {
		log.Errorf("Error while disconnecting from Queue: %s", err.Error())
	}
}

// Add a video ID to VIDEO_SET and to
// VIDEO_WAIT if they are newly found
func SubmitVideoIDs(ids []string) error {
	// Pipeline for querying the status of video IDs
	statusPipe := queue.Pipeline()
	defer statusPipe.Close()

	// Cached cmds
	statusCmds := make([]*redis.IntCmd, len(ids))

	// Queue writes
	for i, id := range ids {
		statusCmds[i] = statusPipe.
			ZAdd(videoSet, redis.Z{float64(time.Now().Unix()), id})
	}

	// Exec writes
	_, err := statusPipe.Exec()
	if err != nil { return err }

	// IDs that get written to VIDEO_WAIT
	var newIDs []interface{}

	// Check if IDs exist
	for i, cmd := range statusCmds {
		// New ID, add to wait queue
		if cmd.Val() == 1 {
			//log.WithField("vid", id).Debug("Found new video")
			newIDs = append(newIDs, ids[i])
		}
	}

	// New IDs, add to wait queue
	queue.LPush(videoWaitQueue, newIDs...)

	return nil
}

// Moves a video from VIDEO_WAIT to VIDEO_WORK
// Possible returns:
//  - "<video-id>", nil: New video ID assigned to this worker
//  - "", nil: No video ID in queue
//  - "", <error>: Error occurred
func GetScheduledVideoIDs(count uint) (ids []string, err error) {
	pipe := queue.Pipeline()
	defer pipe.Close()

	cmds := make([]*redis.StringCmd, count)
	for i := uint(0); i < count; i++ {
		cmds[i] = pipe.RPop(videoWaitQueue)
	}

	// Errors get checked per command
	pipe.Exec()

	// Get IDs from pipe
	for _, cmd := range cmds {
		id, cerr := cmd.Result()

		// End of queue reached
		if id == "" ||
			cerr != nil && cerr.Error() == "redis: nil" { return }

		// Real error
		if cerr != nil { err = cerr; return }

		// Result
		ids = append(ids, id)
	}

	return
}

// Removes a video from VIDEO_WORK
// and places it into VIDEO_WAIT
func FailedVideoID(videoID string) error {
	if err := queue.LPush(videoWaitQueue, videoID).Err();
		err != nil { return err }
	return nil
}

// TODO Recrawl oldest video IDs with "ZPOPMIN VIDEO_SET" & ZADD VIDEO_SET" & "LPUSH VIDEO_WAIT"
