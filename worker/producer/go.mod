module github.com/terorie/yt-mango/worker/producer

go 1.13

replace github.com/terorie/yt-mango => ../../

require (
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/sirupsen/logrus v1.4.2
	github.com/terorie/yt-mango v0.0.0-00010101000000-000000000000
	github.com/ugorji/go v1.1.7 // indirect
	github.com/valyala/fasthttp v1.5.0
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.1.0
)
