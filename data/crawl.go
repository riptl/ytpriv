package data

import "time"

type Crawl struct{
	Video *Video
	VisitedTime time.Time `bson:"visited_time"`
}

type CrawlError struct{
	VideoId string
	Err error
	VisitedTime time.Time `bson:"visited_time"`
}
