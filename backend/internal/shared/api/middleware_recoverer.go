package apicommon

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware recovers from panics and logs them.
func (m *MiddlewareHandler) RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				l := GetLogger(r.Context())
				l.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
