package api

import (
	"APIGateway/comments/pkg/storage"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCommentHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	psgr, err := storage.New(ctx, "postgres://postgres:password@192.168.58.133:5432/comments")
	if err != nil {
		t.Fatal(err)
	}
	api := New(psgr)
	var testBody = []byte(`{"newsID": 1,"content": "Тест"}`)
	req := httptest.NewRequest(http.MethodPost, "/comments/add", bytes.NewBuffer(testBody))
	rr := httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)

	if !(rr.Code == http.StatusCreated) {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
	}
	req = httptest.NewRequest(http.MethodGet, "/comments?news_id=1", nil)
	rr = httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)

	if !(rr.Code == http.StatusOK) {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("не удалось раскодировать ответ сервера: %v", err)
	}

	var data []storage.Comment
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatalf("не удалось раскодировать ответ сервера: %v", err)
	}

	const wantLen = 1
	if len(data) < wantLen {
		t.Fatalf("получено %d записей, ожидалось %d", len(data), wantLen)
	}

	req = httptest.NewRequest(http.MethodPost, "/comments/add", nil)
	rr = httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)
	if !(rr.Code == http.StatusConflict) {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusConflict)
	}
}
