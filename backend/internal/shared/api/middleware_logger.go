package apicommon

import (
	"log/slog"
	"net/http"
	"time"
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

// wrapResponseWriter wraps an http.ResponseWriter and returns both the base
// responseWriter (for accessing statusCode) and the properly-typed wrapper
// that only exposes optional interfaces the underlying ResponseWriter supports.
func wrapResponseWriter(w http.ResponseWriter) (*responseWriter, http.ResponseWriter) {
	base := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	flusher, canFlush := w.(http.Flusher)
	hijacker, canHijack := w.(http.Hijacker)
	pusher, canPush := w.(http.Pusher)

	// Return the appropriate wrapper based on supported interfaces
	switch {
	case canFlush && canHijack && canPush:
		return base, &struct {
			*responseWriter
			http.Flusher
			http.Hijacker
			http.Pusher
		}{base, flusher, hijacker, pusher}
	case canFlush && canHijack:
		return base, &struct {
			*responseWriter
			http.Flusher
			http.Hijacker
		}{base, flusher, hijacker}
	case canFlush && canPush:
		return base, &struct {
			*responseWriter
			http.Flusher
			http.Pusher
		}{base, flusher, pusher}
	case canHijack && canPush:
		return base, &struct {
			*responseWriter
			http.Hijacker
			http.Pusher
		}{base, hijacker, pusher}
	case canFlush:
		return base, &struct {
			*responseWriter
			http.Flusher
		}{base, flusher}
	case canHijack:
		return base, &struct {
			*responseWriter
			http.Hijacker
		}{base, hijacker}
	case canPush:
		return base, &struct {
			*responseWriter
			http.Pusher
		}{base, pusher}
	default:
		return base, base
	}
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
			slog.String("user_agent", r.UserAgent()),
		)

		w.Header().Set(RequestIDHeader, requestID)

		// Store logger and request ID in context
		ctx := WithLogger(r.Context(), reqLogger)

		base, wrapped := wrapResponseWriter(w)

		// Log request start
		start := time.Now()

		reqLogger.Info("request started")

		// Call next handler with enhanced context
		next.ServeHTTP(wrapped, r.WithContext(ctx))

		// Log request completion
		duration := time.Since(start)
		reqLogger.Info("request completed",
			slog.Int("status", base.statusCode), slog.Duration("duration", duration),
		)
	})
}
