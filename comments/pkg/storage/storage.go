// storage/storage.go
package storage

type Comment struct {
	ID       int    `json:"id"`
	NewsID   int    `json:"news_id"`
	ParentID *int   `json:"parent_id,omitempty"`
	Content  string `json:"content"`
	PubTime  int64  `json:"pub_time,omitempty"`
}

type Interface interface {
	AllComments(newsID int) ([]Comment, error)
	AddComment(Comment) error
}
