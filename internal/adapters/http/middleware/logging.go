package middleware

import (
	"net/http"
	"time"

	"github.com/Kartikey2011yadav/mailraven-server/internal/observability"
	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logging creates middleware that logs HTTP requests
func Logging(logger *observability.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := uuid.New().String()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default to 200
			}

			// Log request start
			logger.WithAPI(requestID, r.Method, r.URL.Path).Info("HTTP request started",
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request completion
			duration := time.Since(start)
			logger.WithAPI(requestID, r.Method, r.URL.Path).Info("HTTP request completed",
				"status", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"bytes_written", wrapped.written,
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}
