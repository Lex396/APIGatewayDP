// pkg/middl/middl.go
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
			rID, err := uuid.NewV4()
			if err != nil {
				log.Printf("Ошибка генерации UUID: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			reqID = rID.String()
		}

		req.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

		log.Printf("--> %s %s %s trace_id: %s", req.Method, req.URL.Path, req.RemoteAddr, reqID)

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, req)

		duration := time.Since(start)
		log.Printf("<-- %s %s %d %s (duration: %v) trace_id: %s",
			req.Method,
			req.URL.Path,
			lrw.statusCode,
			http.StatusText(lrw.statusCode),
			duration,
			reqID,
		)
	})
}

type contextKey string

const RequestIDKey contextKey = "requestID"

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
	if ctx == nil {
		return "unknown"
	}
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return "unknown"
}
