package store

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/terorie/yt-mango/viperstruct"
	"github.com/terorie/yt-mango/data"
	"github.com/mongodb/mongo-go-driver/mongo/insertopt"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

var dbClient *mongo.Client
var videos *mongo.Collection

// Mongo database

func ConnectMongo() error {
	// Default config vars
	viper.SetDefault("mongo.conn", "mongodb://127.0.0.1:27017")
	viper.SetDefault("mongo.database", "yt-mango")
	viper.SetDefault("mongo.timeout", 10000)

	var mongoConf struct{
		Conn   string `viper:"mongo.conn"`
		DbName string `viper:"mongo.database"`
		Timeout uint `viper:"mongo.timeout"`
	}

	// Read config
	err := viperstruct.ReadConfig(&mongoConf)
	if err != nil { return err }

	// Create mongo client
	dbClient, err = mongo.NewClientWithOptions(mongoConf.Conn,
		clientopt.SocketTimeout(time.Duration(mongoConf.Timeout) * time.Millisecond))
	if err != nil { return err }

	ctxt := context.Background()

	if err := dbClient.Connect(ctxt);
		err != nil { return err }

	db := dbClient.Database(mongoConf.DbName)
	if db == nil { return errors.New("failed to create database") }

	videos = db.Collection("videos")

	return nil
}

func DisconnectMongo() {
	if err := dbClient.Disconnect(context.Background()); err != nil {
		log.Errorf("Error while disconnecting Mongo: %s", err.Error())
	}
}

func SubmitCrawls(results []data.Crawl) (err error) {
	iResults := make([]interface{}, len(results))
	for i, r := range results {
		iResults[i] = r
	}

	_, err = videos.InsertMany(context.Background(), iResults, insertopt.Ordered(false))
	if err != nil {
		log.Errorf("Uploading crawl of %d videos failed: %s", len(results), err.Error())
	}

	return
}
