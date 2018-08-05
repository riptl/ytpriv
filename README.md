# WIP: yt-mango 💾

> YT metadata extractor inspired by [`youtube-ma` by _CorentinB_][youtube-ma]

__Warning: Very WIP!__ Only `channel dumpurls` and `video detail` work rn

##### Build

Install and compile the Go project with `go get github.com/terorie/yt-mango`!

If you don't have a Go toolchain, grab an executable from the Releases tab

##### Project structure

- _/data_: Data definitions
- _/api_: Abstract API definitions
    - _/apiclassic_: HTML API implementation (parsing using [goquery][goquery])
    - _/apijson_: JSON API implementation (parsing using [fastjson][fastjson])
- _/net_: HTTP utilities (async HTTP implementation)
- _/cmd_: Cobra CLI
- _/util_: I don't have a better place for these

- _/pretty_: (not yet used) Terminal color utilities
- _/worker_: (not yet implemented) worker mode
- _/store_: (not fully implemented) Queue and database
    - _/store/queue.go_: Redis job queue (using [go-redis][go-redis])
    - _/store/db.go_: Mongo main DB (using [the official Mongo driver][mongodb-driver])

 [youtube-ma]: https://github.com/CorentinB/youtube-ma
 [goquery]: https://github.com/PuerkitoBio/goquery
 [fastjson]: https://github.com/valyala/fastjson
 [cobra]: https://github.com/spf13/cobra
 [go-redis]: https://github.com/go-redis/redis
 [mongodb-driver]: https://github.com/mongodb/mongo-go-driver
