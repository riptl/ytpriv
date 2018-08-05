package data

import "time"

type Crawl struct{
	*Video
	VisitedTime time.Time `bson:"visited_time"`
}
