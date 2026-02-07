package apicommon

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

// MiddlewareHandler holds the logger for middleware.
type MiddlewareHandler struct {
	l *slog.Logger
}

// NewMiddlewareHandler creates a new middleware handler.
func NewMiddlewareHandler(l *slog.Logger) *MiddlewareHandler {
	return &MiddlewareHandler{l: l}
}

// RequestIDMiddleware extracts the request ID from the request header or generates a new one
// if it's not present and stores it in the request context.
func (m *MiddlewareHandler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set(RequestIDHeader, requestID)

		// Store request ID in context
		ctx := WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
