# ytwrk 💾

A fast and reliable tool for exporting YouTube data.

__Features__
 - Exports comments
 - Exports live chat
 - Efficient parsing using `fastjson`
 - Detailed access to channels and videos
 - Scripting-friendly, favors stdio and works well with `mlr` and `jq`.
 - Streaming-friendly, produces CSV/TSVs and NDJSON dumps.

If you need a large-scale data export, visit https://files.mine.terorie.dev/yt before starting a crawl.

Only use this tool to the extent permitted by the [YouTube ToS](https://www.youtube.com/static?template=terms).

## Utility Commands

### `video comments [video ...]`

 - Battle-tested exporter
 - Outputs an NDJSON stream
 - With positional arguments:
   Dumps the comments of specified videos
 - Without arguments:
   Reads the video IDs from stdin
 - Example `ytwrk video comments < video_list.txt | jq -r '.content | .[0].text'`
 - Example `ytwrk video comments < video_list.txt | go run ./scripts/commments_to_csv.go | zstdmt -T8 ...`

### `video detail <video>`

 - Takes a video ID, fetches and parses it
 - Outputs some information in JSON
 - Not stable, might change

### `video dump`

 - Outputs an NDJSON stream
 - With positional arguments:
   Dumps the specified videos
 - Without arguments:
   Reads the video IDs from stdin
 - Discovers additional videos through the related section (`-n=<count>` flag)
 - Optionally outputs found but not crawled videos (`-r=<file>` flag)
 - Remembers already crawled videos, so memory usage is O(n)
 - Not stable, might change
 - Example `ytwrk video dump < video_list.txt | nc ...`


### `video dumpraw`

 - Dumps a set of videos in raw tar format
 - Useful for lossless crawls and analyzing data in-post
 - Output is a `.tar` archive stream
 - Reads the video IDs to dump from stdin
 - Example `ytwrk video dumpraw < video_list.txt | tar -xO | jq -r '...'`
 - Example `ytwrk video dumpraw < video_list.txt | zstdmt -T8 ...`

### `video parseraw`

 - Consumes a tar stream produced by `video dumpraw`
 - Produces NDJSON like the `dump` command

### `video live chat`

 - Experimental
 - Exports live chat messages

### `channel dump [channel_id ...]`

 - Get all uploads of channels
 - Prints `channel_id<TAB>video_id` lines
 - With positional arguments:
   Dumps the specified channels
 - Without arguments:
   Reads the channel IDs from stdin
 - Example `ytwrk channel dump < channel_list.txt > video_list.tsv`

### `channel dumpraw`

 - Dumps a set of channels in raw tar format
 - Useful for lossless crawls and analyzing data in-post
 - Output is a `.tar` archive stream
 - Reads the channel IDs to dump from stdin
 - Example `ytwrk channel dumpraw < channel_list.txt | tar -xO | jq -r '...'`
 - Example `ytwrk channel dumpraw < channel_list.txt | zstdmt -T8 ...`

### `playlist videos <playlist_id>`

 - Experimental! Might change without warning
 - Writes all video IDs in the specified playlist to stdout
