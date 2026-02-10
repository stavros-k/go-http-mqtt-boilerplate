package helpers

import (
	"context"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/config"
	"http-mqtt-boilerplate/backend/internal/migrations"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
)

func GetLogger(config *config.Config) *slog.Logger {
	logOptions := slog.HandlerOptions{
		Level:       config.LogLevel,
		ReplaceAttr: utils.SlogReplacer,
	}

	var logHandler slog.Handler = slog.NewJSONHandler(config.LogOutput, &logOptions)
	if config.Generate {
		logHandler = slog.NewTextHandler(config.LogOutput, &logOptions)
	}

	return slog.New(logHandler).With(slog.String("version", utils.GetVersionShort()))
}

func RunMigrations(l *slog.Logger, connString string, dirs ...string) error {
	l.Info("running database migrations")

	// Create migrator with shared + cloud migration directories
	mig, err := migrator.New(l, connString, migrations.GetFS(), dirs...)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := mig.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	l.Info("database migrations completed successfully")

	return nil
}

// NewPgxPool creates a new pgxpool with logging enabled.
func NewPgxPool(ctx context.Context, l *slog.Logger, connString string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Create a child logger for database operations
	dbLogger := l.With(slog.String("component", "database"))

	// Configure tracelog with slog adapter
	config.ConnConfig.Tracer = &tracelog.TraceLog{
		LogLevel: tracelog.LogLevelInfo,
		Logger: tracelog.LoggerFunc(func(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
			attrs := make([]slog.Attr, 0, len(data))
			for k, v := range data {
				attrs = append(attrs, slog.Any(k, v))
			}
			dbLogger.LogAttrs(ctx, traceLogLevelToSlog(level), msg, attrs...)
		}),
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return pool, nil
}

// traceLogLevelToSlog converts tracelog.LogLevel to slog.Level.
func traceLogLevelToSlog(level tracelog.LogLevel) slog.Level {
	switch level {
	case tracelog.LogLevelTrace:
		return slog.LevelDebug - 1
	case tracelog.LogLevelDebug:
		return slog.LevelDebug
	case tracelog.LogLevelInfo:
		return slog.LevelInfo
	case tracelog.LogLevelWarn:
		return slog.LevelWarn
	case tracelog.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
