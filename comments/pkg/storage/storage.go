// storage/storage.go
package storage

type Comment struct {
	ID       int    `json:"id"`
	NewsID   int    `json:"news_id"`
	ParentID *int   `json:"parent_id,omitempty"`
	Content  string `json:"content"`
	PubTime  int64  `json:"pubtime,omitempty"`
}

type Interface interface {
	AllComments(newsID int) ([]Comment, error)
	AddComment(Comment) error
	DeleteComment(id int) error
}
