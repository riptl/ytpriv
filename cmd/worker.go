package cmd

import (
	"os"
	"time"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/terorie/yt-mango/store"
	"github.com/terorie/yt-mango/apis"
	"github.com/terorie/yt-mango/net"
	"github.com/terorie/yt-mango/data"
	"errors"
	"github.com/terorie/yt-mango/api"
)

var fatalErr = errors.New("fatal error, worker must stop")

var Worker = cobra.Command{
	Use: "worker [config file]",
	Short: "Connect to a queue and start archiving",
	Long: "Get work from a Redis queue, start extracting metadata\n" +
		"and upload it to a Mongo database.",
	Args: cobra.MaximumNArgs(1),
	Run: cmdFunc(doWork),
}

func doWork(_ *cobra.Command, args []string) error {
	var overrideFile string

	if len(args) == 1 { overrideFile = args[0] }
	if err := readConfig(overrideFile);
		err != nil { return err }

	if err := store.ConnectQueue();
		err != nil { return err }
	defer store.DisconnectQueue()
	log.Info("Connected to Redis.")

	if err := store.ConnectMongo();
		err != nil { return err }
	defer store.DisconnectMongo()
	log.Info("Connected to Mongo.")

	for {
		videoId, err := store.GetScheduledVideoID()
		if err != nil && err.Error() != "redis: nil" { return err }
		if videoId == "" {
			log.Info("Queue is empty, idling for 10 seconds.")
			time.Sleep(10 * time.Second)
			continue
		}

		// TODO Move video back to wait queue if processing failed

		req := apis.Main.GrabVideo(videoId)
		res, err := net.Client.Do(req)
		if err != nil {
			log.Errorf("Failed to download video \"%s\": %s", videoId, err.Error())
			return fatalErr
		}

		var v data.Video
		v.ID = videoId
		var result interface{}

		next, err := apis.Main.ParseVideo(&v, res)
		if err == api.VideoUnavailable {
			log.Debugf("Video is unavailable: %s", videoId)
			result = data.CrawlError{ uint(api.VideoUnavailable), time.Now() }
		} else if err != nil {
			log.Errorf("Parsing video \"%s\" failed: %s", videoId, err.Error())
			return fatalErr
		} else {
			result = data.Crawl{ &v, time.Now() }
		}

		err = store.SubmitCrawl(result)
		if err != nil {
			log.Errorf("Uploading crawl of video \"%s\" failed: %s", videoId, err.Error())
			return fatalErr
		}

		if len(next) > 0 {
			err = store.SubmitVideoIDs(next)
			if err != nil {
				log.Errorf("Pushing related video IDs of video \"%s\" failed: %s", videoId, err.Error())
				return fatalErr
			}
		}

		err = store.DoneVideoID(videoId)
		if err != nil {
			log.Errorf("Marking video \"%s\" as done failed: %s", videoId, err.Error())
			return fatalErr
		}

		log.Infof("Visited %s.", videoId)
	}

	return nil
}

func readConfig(overrideFile string) error {
	viper.SetConfigType("yaml")
	if overrideFile != "" {
		confFile, err := os.Open(overrideFile)
		if err != nil { return err }
		viper.ReadConfig(confFile)
		return nil
	} else {
		viper.SetConfigName("worker")
		addConfigPaths()
		err := viper.ReadInConfig()
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Warnf("WARNING! NO LOG FILE FOUND: %s\n" +
				"Using default values …", err)
			return nil
		default:
			return err
		}
	}
}
