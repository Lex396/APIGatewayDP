package api

import (
	"APIGateway/aggregator/pkg/logger"
	"APIGateway/comments/pkg/storage"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCommentHandler(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	psgr, err := storage.New(ctx, "postgres://filteruser:fpassword@localhost:5432/comments?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	log, err := logger.NewLogger("comments.log")
	if err != nil {
		t.Fatal(err)
	}

	api := New(psgr, log)

	// === 1. Добавление комментария к новости ===
	var testBody = []byte(`{"news_id": 1, "content": "Комментарий к новости"}`)
	req := httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(testBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusCreated)
		t.Log("Тело ответа:", rr.Body.String())
	}

	// === 2. Получение комментариев по news_id ===
	req = httptest.NewRequest(http.MethodGet, "/comments?news_id=1", nil)
	rr = httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
	}

	b, err := io.ReadAll(rr.Body)
	if err != nil {
		t.Fatalf("не удалось прочитать тело ответа: %v", err)
	}

	var data []storage.Comment
	err = json.Unmarshal(b, &data)
	if err != nil {
		t.Fatalf("не удалось раскодировать ответ сервера: %v", err)
	}

	const wantLen = 1
	if len(data) < wantLen {
		t.Fatalf("получено %d записей, ожидалось минимум %d", len(data), wantLen)
	}

	// === 3. Добавление комментария в ответ на другой ===
	replyBody := []byte(fmt.Sprintf(`{"parent_id": %d, "content": "Ответ на комментарий"}`, data[0].ID))
	req = httptest.NewRequest(http.MethodPost, "/comments", bytes.NewBuffer(replyBody))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	api.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusCreated)
		t.Log("Тело ответа:", rr.Body.String())
	}
}
