package store

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/terorie/yt-mango/viperstruct"
	"github.com/mongodb/mongo-go-driver/bson"
)

var dbClient *mongo.Client
var videos *mongo.Collection

// Mongo database

func ConnectMongo() error {
	// Default config vars
	viper.SetDefault("mongo.host", "mongodb://127.0.0.1:27017")
	viper.SetDefault("mongo.dbName", "yt-mango")

	var mongoConf struct{
		Host string `viper:"mongo.host"`
		User string `viper:"mongo.user,optional"`
		Pass string `viper:"mongo.pass,optional"`
		DbName string `viper:"mongo.dbName"`
	}

	// Read config
	err := viperstruct.ReadConfig(&mongoConf)
	if err != nil { return err }

	// Create mongo client
	dbClient, err = mongo.NewClientWithOptions(
		mongoConf.Host,
		clientopt.Auth(clientopt.Credential{
			Username: mongoConf.User,
			Password: mongoConf.Pass,
		}),
	)
	if err != nil { return err }

	ctxt := context.Background()

	if err := dbClient.Connect(ctxt);
		err != nil { return err }

	db := dbClient.Database(mongoConf.DbName)
	if db == nil { return errors.New("failed to create database") }

	videos = db.Collection("videos")

	// Create indexes on collection
	indexView := videos.Indexes()
	_, err = indexView.CreateMany(ctxt, []mongo.IndexModel{
		// Index video ID
		{ Keys: bson.NewDocument(bson.EC.Int32("video.id", 1)) },
		// Index uploader ID, sort by upload date
		{ Keys: bson.NewDocument(
			bson.EC.Int32("video.uploader_id", 1),
			bson.EC.Int32("video.upload_date", 1),
		)},
		// Index all videos by upload date
		{ Keys: bson.NewDocument(bson.EC.Int32("video.upload_date", 1)) },
		// Index all videos by tags
		{ Keys: bson.NewDocument(bson.EC.Int32("video.tags", 1 ))},
	})
	if err != nil { return err }

	return nil
}

func DisconnectMongo() {
	if err := dbClient.Disconnect(context.Background()); err != nil {
		log.Errorf("Error while disconnecting Mongo: %s", err.Error())
	}
}

func SubmitCrawl(result interface{}) (err error) {
	_, err = videos.InsertOne(context.Background(), result)
	return
}
