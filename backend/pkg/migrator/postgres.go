package migrator

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"http-mqtt-boilerplate/backend/pkg/utils"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgresMigrator struct {
	db      *dbmate.DB
	fs      embed.FS
	connStr string
	l       *slog.Logger
}

// newPostgresMigrator creates a new PostgreSQL migrator. The connection string should be a URL.
// Accepts one embed.FS and multiple migration directory paths (e.g., "shared/migrations", "local/migrations").
func newPostgresMigrator(l *slog.Logger, connStr string, fs embed.FS, migrationDirs ...string) (*postgresMigrator, error) {
	if connStr == "" {
		return nil, errors.New("connection string is required")
	}

	if len(migrationDirs) == 0 {
		return nil, errors.New("at least one migration directory is required")
	}

	// Verify all migration directories exist
	for _, dir := range migrationDirs {
		_, err := fs.ReadDir(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration directory %s: %w", dir, err)
		}
	}

	// Parse the connection string URL for dbmate
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	db := dbmate.New(u)
	db.Strict = true
	db.FS = fs
	db.MigrationsDir = migrationDirs
	db.AutoDumpSchema = false

	l = l.With(slog.String("component", "db-migrator"), slog.String("dialect", "postgres"))
	db.Log = utils.NewSlogWriter(l)

	return &postgresMigrator{
		l:       l,
		db:      db,
		fs:      fs,
		connStr: connStr,
	}, nil
}

// Migrate runs migrations on the PostgreSQL database.
func (m *postgresMigrator) Migrate() error {
	m.l.Info("Migrating database")

	if err := m.db.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// DumpSchema dumps the PostgreSQL database schema to the specified file path.
func (m *postgresMigrator) DumpSchema(filePath string) error {
	m.db.SchemaFile = filePath

	m.l.Info("Dumping schema", slog.String("file", filePath))

	if err := m.db.DumpSchema(); err != nil {
		return fmt.Errorf("failed to dump schema: %w", err)
	}

	return nil
}
