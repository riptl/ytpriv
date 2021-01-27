# ytpriv 💾

A fast tool for exporting YouTube data using their undocumented JSON APIs.

No API keys or logins required, and no stability guarantees given.

If you find it useful, please give it a star!

Please only use this tool to the extent permitted by the [YouTube ToS](https://www.youtube.com/static?template=terms).

## Installation

### From source

Requires a Go 1.14+ toolchain.

Run `go install ./cmd/ytpriv` to install to `$(go env GOPATH)/bin/ytpriv`.

## Features

```
ytpriv [command]
  channel     Scrape a channel
  livestream  Scrape a livestream
  playlist    Scrape a playlist
  video       Scrape a video
```

### Channel

```
ytpriv channel [command]
  overview    Get overview of channel
  videos      Get full list of videos of channel
  videos_page Get videos page of channel
```

### Livestream

```
ytpriv livestream [command]
  chat        Follow the live chat
```

### Playlist

```
ytpriv playlist [command]
  videos      Get full list of videos in playlist
  videos_page Get page of videos of playlist
```

### Video

```
ytpriv video [command]
  comments    Scrape comments of videos
  detail      Get details about a video
```

## Attributions

Developed by [@terorie](https://github.com/terorie)

Using the amazing Go ecosystem including:
- [fasthttp](https://pkg.go.dev/github.com/valyala/fasthttp) and [fastjson](github.com/valyala/fastjson) for fast networking
- [testify](https://pkg.go.dev/github.com/stretchr/testify) regression test helpers
- [cobra](https://pkg.go.dev/github.com/spf13/cobra) for CLI
- [backoff](https://pkg.go.dev/github.com/cenkalti/backoff/v4) for retries
- [zap](https://pkg.go.dev/go.uber.org/zap) logging
