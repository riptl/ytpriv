package cmd

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/terorie/yt-mango/store"
	"github.com/terorie/yt-mango/worker"
	"os"
	"os/signal"
	"syscall"
)

var fatalErr = errors.New("fatal error, worker must stop")
var firstId string

var Worker = cobra.Command{
	Use: "worker [config file]",
	Short: "Connect to a queue and start archiving",
	Long: "Get work from a Redis queue, start extracting metadata\n" +
		"and upload it to a Mongo database.",
	Args: cobra.MaximumNArgs(1),
	Run: cmdFunc(doWork),
}

func init() {
	Worker.Flags().StringVar(&firstId, "first-id", "",
		"First video ID where crawling should begin.\n" +
		"Has to be set if crawling is started on a new queue/database.")
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

	ctxt, cancelFunc := context.WithCancel(context.Background())
	watchExit(cancelFunc)
	worker.Run(ctxt, firstId)

	return nil
}

func watchExit(f context.CancelFunc) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		f()
	}()
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
