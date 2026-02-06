package migrator

import (
	"embed"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"log/slog"
)

// Migrator defines the interface for database migrations and schema operations.
type Migrator interface {
	Migrate() error
	DumpSchema(outputPath string) error
}

// New creates a migrator for the specified dialect.
//
//nolint:ireturn // Returns Migrator interface
func New(l *slog.Logger, d dialect.Dialect, fs embed.FS, connString string) (Migrator, error) {
	if err := d.Validate(); err != nil {
		return nil, fmt.Errorf("invalid dialect: %w", err)
	}

	switch d {
	case dialect.SQLite:
		return newSQLiteMigrator(l, fs, connString)
	case dialect.PostgreSQL:
		return newPostgresMigrator(l, fs, connString)
	default:
		return nil, fmt.Errorf("unsupported dialect: %s", d)
	}
}
