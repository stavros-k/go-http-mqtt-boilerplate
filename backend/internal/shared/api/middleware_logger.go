package apicommon

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter

	statusCode   int
	bytesWritten int64
	written      bool
}

// WriteHeader captures the status code.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// Write captures the status code if WriteHeader was not called.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}

	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += int64(n)
	return n, err
}

// wrapResponseWriter wraps the ResponseWriter to capture status code.
func wrapResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// LoggerMiddleware adds a request-scoped logger to the context and logs requests.
func (m *MiddlewareHandler) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())

		// Create request-scoped logger with context
		reqLogger := m.l.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("protocol", r.Proto),
			slog.String("user_agent", r.UserAgent()),
			slog.Int64("request_bytes", r.ContentLength),
		)

		// Store logger and request ID in context
		ctx := WithLogger(r.Context(), reqLogger)

		wrapped := wrapResponseWriter(w)

		// Log request start
		start := time.Now()

		reqLogger.Info("request started")

		// Call next handler with enhanced context
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Log request completion
		duration := time.Since(start)
		reqLogger.Info("request completed",
			slog.Int("status", wrapped.statusCode),
			slog.Int64("response_bytes", wrapped.bytesWritten),
			slog.Duration("duration", duration),
		)
	})
}
