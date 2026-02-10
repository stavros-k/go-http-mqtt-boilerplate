package migrator

import (
	"embed"
	"errors"
	"log/slog"
)

// Migrator defines the interface for database migrations and schema operations.
type Migrator interface {
	Migrate() error
	DumpSchema(outputPath string) error
}

// New creates a PostgreSQL migrator.
// Accepts one embed.FS and multiple migration directory paths.
//
//nolint:ireturn // Returns Migrator interface
func New(l *slog.Logger, connString string, fs embed.FS, migrationDirs ...string) (Migrator, error) {
	if len(migrationDirs) == 0 {
		return nil, errors.New("at least one migration directory is required")
	}

	return newPostgresMigrator(l, connString, fs, migrationDirs...)
}
