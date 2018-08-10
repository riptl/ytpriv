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

// List with VIDs that currently
// are being crawled
const videoWorkQueue = "VIDEO_WORK"

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
func SubmitVideoID(id string) error {
	// Check against the sorted set
	numAdded, err := queue.ZAdd(videoSet, redis.Z{float64(time.Now().Unix()), id}).Result()
	if err != nil { return err }

	// New ID, add to wait queue
	if numAdded == 1 {
		//log.WithField("vid", id).Debug("Found new video")
		if err := queue.LPush(videoWaitQueue, id).Err();
			err != nil { return err }
	}
	return nil
}

// Moves a video from VIDEO_WAIT to VIDEO_WORK
// Possible returns:
//  - "<video-id>", nil: New video ID assigned to this worker
//  - "", nil: No video ID in queue
//  - "", <error>: Error occurred
func GetScheduledVideoID() (string, error) {
	return queue.RPopLPush(videoWaitQueue, videoWorkQueue).Result()
}

// Removes a video from VIDEO_WORK
// to show that the job is done.
func DoneVideoID(videoID string) error {
	return queue.LRem(videoWorkQueue, -1, videoID).Err()
}

// Removes a video from VIDEO_WORK
// and places it into VIDEO_WAIT
func FailedVideoID(videoID string) error {
	if err := queue.LRem(videoWorkQueue, -1, videoID).Err();
		err != nil { return err }
	if err := queue.LPush(videoWaitQueue, videoID).Err();
		err != nil { return err }
	return nil
}

// TODO Recrawl oldest video IDs with "ZPOPMIN VIDEO_SET" & ZADD VIDEO_SET" & "LPUSH VIDEO_WAIT"
