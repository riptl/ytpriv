package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"github.com/terorie/yt-mango/data"
	"github.com/valyala/fastjson"
	"gopkg.in/confluentinc/confluent-kafka-go.v1/kafka"
)

var (
	esBatch int
	esClient *elasticsearch.Client
	consumer *kafka.Consumer
)

func main() {
	// Initialize and create consumer
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
	consumer, err = kafka.NewConsumer(&config)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	// Kafka topic
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		logrus.Fatal("$KAFKA_TOPIC not set")
	}
	err = consumer.Subscribe(topic, nil)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to subscribe to topic")
	}
	esBatch, err = strconv.Atoi(os.Getenv("ELASTIC_BATCH"))
	if err != nil {
		logrus.WithError(err).Fatal("Invalid $ELASTIC_BATCH")
	}

	// Connect to Elastic
	esClient, err = elasticsearch.NewClient(elasticsearch.Config{
		Transport: &http.Transport{
			ResponseHeaderTimeout: 30 * time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
		},
	})
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to ElasticSearch")
	}

	events := consumer.Events()

	var b batch
	for ev := range events {
		switch e := ev.(type) {
		case kafka.AssignedPartitions:
			fmt.Fprintf(os.Stderr, "%% %v\n", e)
			consumer.Assign(e.Partitions)
		case kafka.RevokedPartitions:
			fmt.Fprintf(os.Stderr, "%% %v\n", e)
			consumer.Unassign()
		case *kafka.Message:
			b.append(e.Value)
		case kafka.PartitionEOF:
			fmt.Printf("%% Reached %v\n", e)
		case kafka.Error:
			// Errors should generally be considered as informational, the client will try to automatically recover
			fmt.Fprintf(os.Stderr, "%% Error: %v\n", e)
		}
	}
}

type batch struct {
	buf bytes.Buffer
	enc *json.Encoder
	lines int
}

type elasticScriptParams struct {
	CreatedBefore int64  `json:"created_before"`
	CreatedAfter  int64  `json:"created_after"`
	CrawledAt     int64  `json:"crawled_at"`
	LikeCount     uint64 `json:"likes"`
	ReplyCount    uint64 `json:"replies"`
}

type elasticUpsert struct {
	VideoID        string          `json:"video_id"`
	AuthorID       string          `json:"author_id"`
	ByChannelOwner bool            `json:"by_channel_owner,omitempty"`
	ParentID       string          `json:"parent_id,omitempty"`
	CreatedBefore  int64           `json:"created_before"`
	CreatedAfter   int64           `json:"created_after"`
	Author         string          `json:"author"`
	Content        json.RawMessage `json:"content"`
	Edited         bool            `json:"edited,omitempty"`
	FirstSeen      int64           `json:"first_seen"`
	CrawledAt      int64           `json:"crawled_at"`
	LikeCount      uint64          `json:"likes"`
	ReplyCount     uint64          `json:"replies"`
}

func (b *batch) append(item []byte) {
	var c data.Comment
	err := json.Unmarshal(item, &c)
	if err != nil {
		logrus.WithError(err).Error("Failed to deserialize item")
		return
	}
	esScriptParams := elasticScriptParams{
		CreatedBefore: c.CreatedBefore,
		CreatedAfter:  c.CreatedAfter,
		CrawledAt:     c.CrawledAt,
		LikeCount:     c.LikeCount,
		ReplyCount:    c.ReplyCount,
	}
	esUpsert := elasticUpsert{
		VideoID:        c.VideoID,
		AuthorID:       c.AuthorID,
		ByChannelOwner: c.ByChannelOwner,
		ParentID:       c.ParentID,
		CreatedBefore:  c.CreatedBefore,
		CreatedAfter:   c.CreatedAfter,
		Author:         c.Author,
		Content:        c.Content,
		Edited:         c.Edited,
		FirstSeen:      c.CrawledAt,
		CrawledAt:      c.CrawledAt,
		LikeCount:      c.LikeCount,
		ReplyCount:     c.ReplyCount,
	}

	var reqLine struct{
		Update struct{
			ID string `json:"_id"`
		} `json:"update"`
	}
	reqLine.Update.ID = c.ID

	b.enc = json.NewEncoder(&b.buf)
	err = b.enc.Encode(&reqLine)
	if err != nil {
		panic(err)
	}

	var dataLine struct {
		Script struct {
			ID string `json:"id"`
			Params elasticScriptParams `json:"params"`
		} `json:"script"`
		Upsert elasticUpsert `json:"upsert"`
	}
	dataLine.Script.ID = "comment-upsert"
	dataLine.Script.Params = esScriptParams
	dataLine.Upsert = esUpsert

	err = b.enc.Encode(&dataLine)
	if err != nil {
		panic(err)
	}

	b.lines++
	if b.lines >= esBatch {
		b.flush()
		_, err = consumer.Commit()
		if err != nil {
			logrus.WithError(err).Warn("Failed to commit Kafka offsets")
		}
	}
}

func (b *batch) flush() {
	logrus.Info("Flushing")
	reqBuf := b.buf.Bytes()
	_ = backoff.Retry(func() error {
		res, err := esClient.Bulk(bytes.NewReader(reqBuf), esClient.Bulk.WithIndex("comment-index"))
		if err != nil {
			logrus.WithError(err).Error("Failed to upsert batch")
			return err
		}
		if res.IsError() {
			logrus.Errorf("Failed to upsert batch: HTTP %d", res.StatusCode)
			return err
		}
		resBuf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var p fastjson.Parser
		root, err := p.ParseBytes(resBuf)
		if err != nil {
			return err
		}
		items := root.GetArray("items")
		errorCount := 0
		for _, item := range items {
			update := item.Get("update")
			status := update.GetInt("status")
			if status >= 300 {
				errorCount++
				logrus.WithField("error", update.String()).
					WithField("id", string(item.GetStringBytes("_id"))).
					Error("Failed to upsert item")
			}
		}
		logrus.WithFields(logrus.Fields{
			"items": len(items),
			"errors": errorCount,
		}).Info("Upserted batch")
		return nil
	}, backoff.NewExponentialBackOff())
	b.lines = 0
	b.buf.Reset()
}
