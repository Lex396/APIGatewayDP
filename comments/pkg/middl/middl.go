package middl

import (
	"context"
	"github.com/gofrs/uuid"
	"log"
	"net/http"
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
		reqID := req.URL.Query().Get("request_id")
		if reqID == "" {
			rID, _ := uuid.NewV4()
			reqID = rID.String()
		}
		req.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, req)
		statusCode := lrw.statusCode
		log.Printf("<-- client ip: %s, method: %s, url: %s, status code: %d %s, trace id: %s",
			req.RemoteAddr, req.Method, req.URL.Path, statusCode, http.StatusText(statusCode), reqID)
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
