# WIP: yt-mango 💾

![build status](https://travis-ci.org/terorie/yt-mango.svg?branch=master)
> YT metadata extractor inspired by [`youtube-ma` by _CorentinB_][youtube-ma]

##### Get it

Grab an executable from the Releases tab

Or install and compile the Go project
with `go get github.com/terorie/yt-mango` for a newer build!

##### Usage

- Getting JSON info about a video:
  ```json
  # ./yt-mango video detail https://www.youtube.com/watch?v=imooXqWLOfA
  {
      "id": "imooXqWLOfA",
      "url": "https://www.youtube.com/watch?v=imooXqWLOfA",
      "duration": 264,
      "tags": [
              "nimiq",
              "blockchain",
              "crypto",
              "gource"
      ],
      …
  }
  ```
- [Farming videos (~2000 videos per second)](worker.md)

##### Project structure

- _/data_: Data definitions
- _/api_: Shared functions and abstract API definitions
- _/apis_: API implementations
    - _/apiclassic_: HTML API implementation (parsing using [goquery][goquery])
    - _/apijson_: JSON API implementation (parsing using [fastjson][fastjson])
- _/net_: HTTP utilities (async HTTP implementation)
- _/cmd_: [Cobra][cobra] CLI
- _/store_: Queue and database
    - _/store/queue.go_: Redis job queue (using [go-redis][go-redis])
    - _/store/db.go_: Mongo main DB (using [the official Mongo driver][mongodb-driver])
- _/worker_ Worker mode

- _/pretty_: (not yet used) Terminal color utilities

 [youtube-ma]: https://github.com/CorentinB/youtube-ma
 [goquery]: https://github.com/PuerkitoBio/goquery
 [fastjson]: https://github.com/valyala/fastjson
 [cobra]: https://github.com/spf13/cobra
 [viper]: https://github.com/spf13/viper
 [go-redis]: https://github.com/go-redis/redis
 [mongodb-driver]: https://github.com/mongodb/mongo-go-driver
