package store

import (
	"log"
	"time"
	"github.com/go-redis/redis"
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

// Timeout for blocking functions
const timeout = 30 * time.Second

var client redis.Client

// Redis queue

// Adds video IDs to VIDEO_SET and to
// VIDEO_WAIT if they are newly found
func SubmitVideos(ids []string) error {
	for _, id := range ids {
		// Check against the sorted set
		numAdded, err := client.ZAdd(videoSet, redis.Z{float64(time.Now().Unix()), id}).Result()
		if err != nil { return err }

		// New ID, add to wait queue
		if numAdded == 1 {
			log.Printf("Added new video ID \"%s\" to wait queue.")
			if err := client.LPush(videoWaitQueue, id).Err();
				err != nil { return err }
		}
	}
	return nil
}

// Moves a video from VIDEO_WAIT to VIDEO_SET
// Possible returns:
//  - "<video-id>", nil: New video ID assigned to this worker
//  - "", nil: No video ID in queue
//  - "", <error>: Error occurred
func GetScheduledVideo() (string, error) {
	return client.BRPopLPush(videoWaitQueue, videoWorkQueue, timeout).Result()
}

// TODO Recrawl oldest video IDs with "ZPOPMIN VIDEO_SET" & ZADD VIDEO_SET" & "LPUSH VIDEO_WAIT"
