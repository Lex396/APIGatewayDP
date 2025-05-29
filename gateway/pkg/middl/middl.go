package middl

import (
	"context"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
)

type contextKey string

const RequestIDKey = contextKey("requestID")

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
		reqID := req.Header.Get("X-Request-ID")
		if reqID == "" {
			rID, _ := uuid.NewV4()
			reqID = rID.String()
		}
		req.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Request-ID")

		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, req)

		log.Printf("[Gateway] IP: %s, %s %s, status %d, request_id: %s",
			req.RemoteAddr, req.Method, req.URL.Path, lrw.statusCode, reqID)
	})
}

func GetRequestID(ctx context.Context) string {
	val, ok := ctx.Value(RequestIDKey).(string)
	if !ok {
		return "unknown"
	}
	return val
}
