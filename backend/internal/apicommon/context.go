package apicommon

import (
	"context"
	"log/slog"
)

type contextKey struct{}

//nolint:gochecknoglobals // Context keys must be package-level variables
var (
	loggerKey    = contextKey{}
	requestIDKey = contextKey{}
)

const zeroUUIDForContext = "00000000-0000-0000-0000-000000000000"

// WithLogger adds a request-scoped logger to the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLogger retrieves the request-scoped logger from context.
func GetLogger(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}

	return slog.Default() // Fallback (should never happen)
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}

	return zeroUUIDForContext
}
