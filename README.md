# yt-mango 💾

![build status](https://travis-ci.org/terorie/yt-mango.svg?branch=master)

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
- Download a collection of related videos:
  ```
  # ./yt-mango video dump 1000 dir imooXqWLOfA -c 64
  …
  INFO[0030] Downloaded 1000 videos in 30s!
  # ls -1U dir | wc -l
  1000
  ```

##### Project structure

- _/data_: Data definitions
- _/api_: Shared functions and abstract API definitions
- _/apis_: API implementations
    - _/apiclassic_: HTML API implementation (parsing using [goquery][goquery])
    - _/apijson_: JSON API implementation (parsing using [fastjson][fastjson])
- _/net_: HTTP utilities
- _/cmd_: [Cobra][cobra] CLI

 [goquery]: https://github.com/PuerkitoBio/goquery
 [fastjson]: https://github.com/valyala/fastjson
 [cobra]: https://github.com/spf13/cobra
 [viper]: https://github.com/spf13/viper
