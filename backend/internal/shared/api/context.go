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

// WithLogger adds a request-scoped logger to the context.
func WithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// GetLoggerFromContext retrieves the request-scoped logger from context.
func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	l := GetLoggerFromContextOrNil(ctx)
	if l != nil {
		return l
	}

	return slog.Default() // Fallback (should never happen)
}

// GetLoggerFromContextOrNil retrieves the request-scoped logger from context or returns nil if not set.
func GetLoggerFromContextOrNil(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}

	return nil
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestIDFromContext retrieves the request ID from context.
func GetRequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}

	return zeroUUID
}
