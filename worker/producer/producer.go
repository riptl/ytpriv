package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"

	"github.com/cenkalti/backoff"
	"github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/net"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

func init() {
	net.Client.MaxConnsPerHost = 250
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
	})
}

func main() {
	// Initialize and create producer
	configPath := os.Getenv("KAFKA_CONFIG")
	if configPath == "" {
		logrus.Fatal("$KAFKA_CONFIG not set")
	}
	configBuf, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to read Kafka config")
	}
	var config kafka.ConfigMap
	err = json.Unmarshal(configBuf, &config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse Kafka config JSON")
	}
	producer, err := kafka.NewProducer(&config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Kafka topic
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		logrus.Fatal("$KAFKA_TOPIC not set")
	}
	topicPartition := kafka.TopicPartition{
		Topic:     &topic,
		Partition: kafka.PartitionAny,
	}

	// Listen for Ctrl+C
	ctx, cancel := context.WithCancel(context.Background())
	exitC := make(chan os.Signal)
	signal.Notify(exitC, os.Interrupt)

	// Single producer queue
	messages := make(chan []byte)

	// Start worker scheduler
	var scheduler Scheduler
	err = scheduler.Init(ctx, messages)
	if err != nil {
		logrus.Fatal("Failed to initialize scheduler")
	}
	go scheduler.Run()

	// Stream messages
	for {
		select {
		case <-exitC:
			cancel()
			goto quit
		case msgBuf := <-messages:
			msg := kafka.Message{
				TopicPartition: topicPartition,
				Value: msgBuf,
			}
			err = backoff.Retry(func() error {
				err := producer.Produce(&msg, nil)
				if e, ok := err.(kafka.Error); ok {
					if e.Code() == kafka.ErrQueueFull {
						logrus.WithError(err).Warn("Queue is full, throttling")
						return err
					} else {
						return backoff.Permanent(err)
					}
				} else if err != nil{
					return backoff.Permanent(err)
				} else {
					return nil
				}
			}, backoff.NewExponentialBackOff())
			if err != nil {
				logrus.WithError(err).Error("Failed to produce message")
			}
		}
	}
	quit:

	// Stop scheduler
	scheduler.Close()

	// Shutdown with timeout
	closeTimeout, err := strconv.Atoi(os.Getenv("KAFKA_CLOSE_TIMEOUT"))
	if err != nil {
		logrus.WithError(err).Error("Failed to parse $KAFKA_CLOSE_TIMEOUT")
		closeTimeout = 3000
	}
	producer.Flush(closeTimeout)
}
