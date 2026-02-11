package apicommon

import (
	"http-mqtt-boilerplate/backend/internal/shared/types"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and logs them.
func (m *MiddlewareHandler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nolint:contextcheck // Context accessed from closure is safe in defer recover
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestIDFromContext(r.Context())

				l := GetLoggerFromContextOrNil(r.Context())
				if l == nil {
					l = m.l
				}

				l.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)

				// Respond with a generic error message to avoid leaking internal details
				RespondJSON(w, r, http.StatusInternalServerError, &types.ErrorResponse{
					RequestID: requestID,
					Message:   "Internal Server Error",
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
