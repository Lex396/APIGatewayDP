package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHandler(t *testing.T) {
	api := New()

	tests := []struct {
		name       string
		body       []byte
		wantStatus int
	}{
		{
			name:       "clean comment",
			body:       []byte(`{"content": "Это нормальный комментарий"}`),
			wantStatus: http.StatusOK,
		},
		{
			name:       "contains forbidden word qwerty",
			body:       []byte(`{"content": "это qwerty сообщение"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "contains forbidden word йцукен",
			body:       []byte(`{"content": "это йцукен слово"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "contains forbidden word zxvbnm",
			body:       []byte(`{"content": "плохое zxvbnm слово"}`),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       []byte(`{"invalid":`),
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodPost, "/check", bytes.NewBuffer(tt.body))
		rr := httptest.NewRecorder()
		api.Router().ServeHTTP(rr, req)

		if rr.Code != tt.wantStatus {
			t.Errorf("[%s] expected status %d, got %d", tt.name, tt.wantStatus, rr.Code)
		}
	}
}
