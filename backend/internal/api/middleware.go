package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter

	statusCode int
	written    bool
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

	return rw.ResponseWriter.Write(b)
}

func (s *Handler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store request ID in context
		ctx := WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggerMiddleware adds a request-scoped logger to the context and logs requests.
func (s *Handler) LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		// Create request-scoped logger with context
		reqLogger := s.l.With(
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		w.Header().Set(RequestIDHeader, requestID)

		// Store logger and request ID in context
		ctx := WithLogger(r.Context(), reqLogger)

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Log request start
		start := time.Now()

		reqLogger.Info("request started")

		// Call next handler with enhanced context
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Log request completion
		duration := time.Since(start)
		reqLogger.Info("request completed",
			slog.Int("status", wrapped.statusCode), slog.Duration("duration", duration),
		)
	})
}
