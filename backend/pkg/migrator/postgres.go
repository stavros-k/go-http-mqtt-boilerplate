package migrator

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"net/url"
	"os"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/amacneil/dbmate/v2/pkg/dbutil"
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
func newPostgresMigrator(l *slog.Logger, fs embed.FS, connStr string) (*postgresMigrator, error) {
	if connStr == "" {
		return nil, errors.New("connection string is required")
	}

	_, err := fs.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Parse the connection string URL for dbmate
	u, err := url.Parse(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	db := dbmate.New(u)
	db.Strict = true
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
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

	// read the schema file
	schemaBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	schemaBytes, err = dbutil.StripPsqlMetaCommands(schemaBytes)
	if err != nil {
		return fmt.Errorf("failed to strip psql meta commands: %w", err)
	}

	schema := string(bytes.TrimSpace(schemaBytes)) + "\n"

	if err := os.WriteFile(filePath, []byte(schema), 0o600); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	return nil
}
