module github.com/terorie/yt-mango/worker/consumer

go 1.13

replace github.com/terorie/yt-mango => ../../

require (
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/elastic/go-elasticsearch/v7 v7.3.0
	github.com/sirupsen/logrus v1.4.2
	github.com/terorie/yt-mango v0.0.0-00010101000000-000000000000
	github.com/valyala/fastjson v1.4.1
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.1.0
)
