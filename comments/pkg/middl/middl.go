package middl

import (
	"context"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
	"time"
)

type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

type contextKey string

const RequestIDKey = contextKey("requestID")

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {

	return &LoggingResponseWriter{w, http.StatusOK}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Middle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		reqID := req.URL.Query().Get("request_id")
		if reqID == "" {
			rID, _ := uuid.NewV4()
			reqID = rID.String()
		}
		req.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)

		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, req)

		log.Printf(
			"<-- time: %s, client ip: %s, method: %s, url: %s, status code: %d %s, trace id: %s",
			start.Format(time.RFC3339),
			req.RemoteAddr,
			req.Method,
			req.URL.Path,
			lrw.statusCode,
			http.StatusText(lrw.statusCode),
			reqID,
		)
	})
}

func WithRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = "unknown"
		}
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetRequestID(ctx context.Context) string {
	val, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return val
}
