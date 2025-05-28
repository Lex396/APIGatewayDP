package api

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/aggregator/pkg/storage"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

type MockStorage struct {
	posts []storage.Post
	err   error
}

func (m *MockStorage) SavePost(post storage.Post, logInstance *logger.Logger) error {
	return nil
}

func (m *MockStorage) GetLastPosts(limit int, page int) ([]storage.Post, storage.Pagination, error) {
	if m.err != nil {
		return nil, storage.Pagination{}, m.err
	}

	start := (page - 1) * limit
	end := start + limit
	if start > len(m.posts) {
		return []storage.Post{}, storage.Pagination{}, nil
	}
	if end > len(m.posts) {
		end = len(m.posts)
	}

	paginated := m.posts[start:end]

	pagination := storage.Pagination{
		NumOfPages: 1,
		Page:       1,
		Limit:      15,
	}

	return paginated, pagination, nil
}

func (m *MockStorage) GetPostByID(id int) (storage.Post, error) {
	for _, post := range m.posts {
		if post.ID == id {
			return post, nil
		}
	}
	return storage.Post{}, errors.New("post not found")
}

func (m *MockStorage) CountPostsByTitle(title string) (int, error) {
	return len(m.posts), nil
}

func (m *MockStorage) GetPostsByTitle(search string, limit, offset int) ([]storage.Post, storage.Pagination, error) {
	if m.err != nil {
		return nil, storage.Pagination{}, m.err
	}
	return m.posts, storage.Pagination{
		Page:       offset,
		Limit:      limit,
		NumOfPages: 1,
	}, nil
}

func newTestLogger() *logger.Logger {
	logInstance, _ := logger.NewLogger("test.log")
	return logInstance
}

func newTestAPI(db storage.StorageInterface) *API {
	return &API{DB: db, Logger: newTestLogger()}
}

func setupTestTemplate() {
	os.MkdirAll("web", os.ModePerm)
	os.WriteFile("web/index.html", []byte("{{range .Posts}}<h1>{{.Title}}</h1>{{end}}"), 0644)
}

func TestMain(m *testing.M) {
	setupTestTemplate()
	os.Exit(m.Run())
}

func TestAPI_postsHandler(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		mockDB     *MockStorage
		wantStatus int
	}{
		{"Valid limit", "/news/3", &MockStorage{posts: generateMockPosts(3)}, http.StatusOK},
		{"Invalid limit", "/news?limit=abc", &MockStorage{posts: generateMockPosts(3)}, http.StatusBadRequest},
		{"DB error", "/news/5", &MockStorage{err: errors.New("ошибка БД")}, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newTestAPI(tt.mockDB)
			req := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			api.postsHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Ожидался статус %d, получено: %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestTrimContent(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"Это тестовое сообщение", 10, "Это тестов..."},
		{"Короткий текст", 20, "Короткий текст"},
		{"Привет, мир!", 5, "Приве..."},
	}

	for _, tt := range tests {
		got := trimContent(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("Ожидалось: %s, получено: %s", tt.want, got)
		}
	}
}

func generateMockPosts(count int) []storage.Post {
	var posts []storage.Post
	for i := 1; i <= count; i++ {
		posts = append(posts, storage.Post{
			ID:      i,
			Title:   "Новость " + strconv.Itoa(i),
			Content: "Контент новости " + strconv.Itoa(i),
			PubTime: time.Now(),
			Link:    "http://example.com/" + strconv.Itoa(i),
		})
	}
	return posts
}

func TestAPI_NewsDetailedHandler(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		mockDB     *MockStorage
		wantStatus int
	}{
		{
			name:       "Valid ID",
			query:      "/news/search?id=2",
			mockDB:     &MockStorage{posts: generateMockPosts(5)},
			wantStatus: http.StatusOK,
		},
		{
			name:       "Missing ID",
			query:      "/news/search",
			mockDB:     &MockStorage{posts: generateMockPosts(5)},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid ID",
			query:      "/news/search?id=abc",
			mockDB:     &MockStorage{posts: generateMockPosts(5)},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Not Found",
			query:      "/news/search?id=999",
			mockDB:     &MockStorage{posts: generateMockPosts(5)},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := newTestAPI(tt.mockDB)
			req := httptest.NewRequest("GET", tt.query, nil)
			w := httptest.NewRecorder()

			api.newsDetailedHandler(w, req)

			resp := w.Result()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Ожидался статус %d, получено: %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}
