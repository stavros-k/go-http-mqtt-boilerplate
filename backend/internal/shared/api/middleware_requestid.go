package apicommon

import (
	"http-mqtt-boilerplate/backend/pkg/utils"
	"net/http"

	"github.com/google/uuid"
)

// RequestIDMiddleware extracts the request ID from the request header or generates a new one
// if it's not present and stores it in the request context.
func (m *MiddlewareHandler) RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get or generate request ID
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			uuid, err := uuid.NewV7()
			if err != nil {
				m.l.Error("failed to generate request ID", utils.ErrAttr(err))
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				return
			}
			requestID = uuid.String()
		}

		w.Header().Set(RequestIDHeader, requestID)

		// Store request ID in context
		ctx := WithRequestID(r.Context(), requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
