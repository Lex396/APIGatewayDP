package api

import (
	"APIGateway/gateway/config"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockServer создает простой mock-сервер с указанным ответом и кодом
func mockServer(status int, responseBody interface{}) *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		if responseBody != nil {
			json.NewEncoder(w).Encode(responseBody)
		}
	})
	return httptest.NewServer(handler)
}

func TestHandleGetNews(t *testing.T) {
	newsResp := map[string]interface{}{
		"items": []map[string]string{{"title": "News1"}, {"title": "News2"}},
	}
	newsSrv := mockServer(http.StatusOK, newsResp)
	defer newsSrv.Close()

	cfg := &config.Config{}
	a := New(cfg, newsSrv.URL[len("http://localhost"):], "", "")
	req := httptest.NewRequest("GET", "/news", nil)
	w := httptest.NewRecorder()

	a.Router().ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !bytes.Contains(body, []byte("News1")) {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestHandleGetNewsByID(t *testing.T) {
	newsResp := map[string]interface{}{"id": 1, "title": "Test News"}
	commentsResp := []map[string]interface{}{
		{"author": "Alice", "text": "Great!"},
	}

	newsSrv := mockServer(http.StatusOK, newsResp)
	commentsSrv := mockServer(http.StatusOK, commentsResp)

	defer newsSrv.Close()
	defer commentsSrv.Close()

	cfg := &config.Config{}
	a := New(cfg, newsSrv.URL[len("http://localhost"):], "", commentsSrv.URL[len("http://localhost"):])

	req := httptest.NewRequest("GET", "/news/1", nil)
	w := httptest.NewRecorder()

	a.Router().ServeHTTP(w, req)
	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}
	if !bytes.Contains(body, []byte("Test News")) || !bytes.Contains(body, []byte("Alice")) {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestHandlePostComment(t *testing.T) {
	censorSrv := mockServer(http.StatusOK, map[string]string{"text": "cleaned comment"})
	commentsSrv := mockServer(http.StatusCreated, map[string]string{"status": "ok"})

	defer censorSrv.Close()
	defer commentsSrv.Close()

	cfg := &config.Config{}
	a := New(cfg, "", censorSrv.URL[len("http://localhost"):], commentsSrv.URL[len("http://localhost"):])

	comment := map[string]string{"text": "badword"}
	body, _ := json.Marshal(comment)
	req := httptest.NewRequest("POST", "/news/1/comments", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	a.Router().ServeHTTP(w, req)

	resp := w.Result()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resp.StatusCode)
	}
	if !bytes.Contains(respBody, []byte("ok")) {
		t.Errorf("unexpected body: %s", string(respBody))
	}
}
