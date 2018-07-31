package data

type Channel struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Paid bool `json:"paid"`
	Thumbnail string `json:"thumbnail"`
}
