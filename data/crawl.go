package data

import "time"

type Crawl struct{
	Video *Video `bson:"video"`
	VisitedTime time.Time `bson:"visited_time"`
	CrawlerName string `bson:"found_by"`
}

type CrawlError struct{
	VideoId string
	Err error
	VisitedTime time.Time `bson:"visited_time"`
}
