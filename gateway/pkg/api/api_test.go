package api

import (
	"APIGateway/gateway/config"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPI_endpoints(t *testing.T) {
	cfg := &config.Config{
		Censor: config.Censor{
			AdrPort: ":8081",
		},
		Comments: config.Comments{
			AdrPort: ":8082",
		},
		News: config.News{
			AdrPort: ":8083",
		},
		Gateway: config.Gateway{
			AdrPort: ":8080",
		},
	}

	api := New(cfg, "8083", "8081", "8082")

	t.Run("Test /news endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news", nil)
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("Test /news/latest endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/latest", nil)
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
		}
	})

	t.Run("Test /news/search endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/news/search?id=2", nil)
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
		}
	})

	var testBody1 = []byte(`{"newsID": 3,"content": "Тест qwerty "}`)
	var testBody2 = []byte(`{"newsID": 3,"content": "Тест ups "}`)
	var testBody3 = []byte(`{"id": 3}`)

	t.Run("Test invalid comment add", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/comments/add", bytes.NewBuffer(testBody1))
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusBadRequest)
		}
	})

	t.Run("Test valid comment add", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/comments/add", bytes.NewBuffer(testBody2))
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusCreated)
		}
	})

	t.Run("Test comment delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/comments/del", bytes.NewBuffer(testBody3))
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("код неверен: получили %d, а хотели %d", rr.Code, http.StatusOK)
		}
	})
}
