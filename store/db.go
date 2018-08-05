package store

import (
	"time"
	"context"
	"errors"
	"log"
	"github.com/spf13/viper"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/terorie/yt-mango/data"
	"github.com/terorie/yt-mango/viperstruct"
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

	if err := dbClient.Connect(context.TODO());
		err != nil { return err }

	db := dbClient.Database(mongoConf.DbName)
	if db == nil { return errors.New("failed to create database") }

	videos = db.Collection("videos")

	return nil
}

func DisconnectMongo() {
	if err := dbClient.Disconnect(context.Background()); err != nil {
		log.Fatalf("Error while disconnecting Mongo: %s", err.Error())
	}
}

func SubmitCrawl(video *data.Video, crawlTime time.Time) (err error) {
	_, err = videos.InsertOne(context.Background(), video)
	return
}
