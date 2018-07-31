# Ideas/Questions for future code

> Brainstorming here :P

### MongoDB

* Main database for crawled content
* YouTube is mostly non-relational
  (except channels ↔ videos)
* Users can change videos (title, etc.):
  Support for multiple crawls needed
    * In one document as array?
      Like `{videoid: xxx, crawls: []…`
        * Pros: Easy history query
        * Cons: (Title) indices might be harder to maintain
    * Or as separate documents?
      Like `{videoid: xxx, crawldate: …`
        * Pros: Race conditions less likely
        * Cons: Duplicates more likely?
    * Avoiding duplicates?
        * If the user hasn't changed video metadata,
          crawling it again is a waste of disk space
        * Rescan score: Should a video be rescanned?
            * Viral videos should be crawled more often
            * New videos shouldn't be instantly crawled again
            * Very old videos are unlikely to change
            * Maybe focus on views per week
            * Machine learning?
        * Hashing data from crawls to detect changes?
            * Invalidates old data on API upgrade
            * Could be used as an index tho
* Live data
    * like views/comments/subscribers per day
    * vs more persistent data: Title/Description/video Formats
    * Are they worth crawling
* Additional data
    * Like subtitles and annotations
    * Need separate crawls
    * Not as important as main data

### Types of bot

* __Discover bots__
    * Find and push new video IDs to the queue
    * Monitor channels for new content
    * Discover new videos 
* __Maintainer bots__
    * Occasionally look at the database and 
      push backups/freezes to drive
    * Decide which old video IDs to re-add to the queue
* __Worker bots__
    * Get jobs from the Redis queue and crawl YT
    * Remove processed entries from the queue

### Redis queue

* A redis queue lists video IDs that have been
  discovered, but not crawled
* Discover bots bots push IDs if they find new ones
    * Implement queue priority?
* Maintainer bots push IDs if they likely need rescans
* States of queued items
    1. _Queued:_ Processing required
       (no worker bot picked them up yet)
    2. _Assigned:_ Worker claimed ID and processes it.
       If the worker doesn't mark the ID as done in time
       it gets tagged back as _Queued_ again
       (should be hidden from other workers)
    3. _Done:_ Worker submitted crawl to the database
       (can be deleted from the queue)
* Single point of failure
    * Potentially needs ton of RAM
        * 800 M IDs at 100 bytes per entry = 80 GB
    * Shuts down entire crawl system on failure
    * Persistence: A crash can use all discovered IDs
* Alternative implementations
    * SQLite in-memory?
