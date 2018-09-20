##### Worker mode

A special worker mode uses a [Redis](https://redis.io/)
job queue to manage tasks (uncrawled YouTube videos).
The metadata of each video will be uploaded to a central
[Mongo](https://www.mongodb.com/) database afterwards.

While the job queue and DB are central, you can run
as many workers as you want!
Tested scaling up to ~2k video visits per second.

The process basically consists of three steps and gets repeated possibly forever:
 1. Get a random video page from the job queue.
 2. Visit the video page, and extract info (title, description, etc.)
 3. Get the recommended videos on the page and place new ones into the job queue.

This is essentially a distributed
[Breadth-First-Search](https://en.wikipedia.org/wiki/Breadth-first_search)
on the entire YouTube page!

Test it out yourself!
Mac:
```
$ brew install redis mongo
$ brew services launch redis mongo
```

Debian/Ubuntu: (mongodb-org 4.0 recommended!)
```
# apt install redis mongodb
# systemctl start redis
# systemctl start mongodb
```

Then, start yt-mango: 
```
$ yt-mango worker -a json config.yml --first-id 5Erj9y4D3iY
```

![yt-mango worker output](doc/working.png)

By default, yt-mango tries to connect to Redis and Mongo on localhost.
This can be changed using a config file (example in `/example.yml`).
