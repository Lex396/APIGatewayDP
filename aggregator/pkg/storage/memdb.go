package storage

import (
	"APIGateway/aggregator/pkg/logger"
	"sync"
	"time"
)

type MemDB struct {
	mu    sync.Mutex
	posts []Post
}

func NewMemDB() *MemDB {
	return &MemDB{}
}

// SavePost сохраняет пост в память
func (m *MemDB) SavePost(post Post, logInstance *logger.Logger) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, p := range m.posts {
		if p.Link == post.Link {
			return nil
		}
	}

	post.ID = len(m.posts)
	m.posts = append(m.posts, post)
	return nil
}

// GetLastPosts возвращает последние n постов
func (m *MemDB) GetLastPosts(limit int, fromDate time.Time) ([]Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []Post
	for i := len(m.posts) - 1; i >= 0; i-- {
		if m.posts[i].PubTime.After(fromDate) {
			continue
		}
		result = append(result, m.posts[i])
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}
