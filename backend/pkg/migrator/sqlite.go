package migrator

import (
	"embed"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"net/url"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteMigrator struct {
	db      *dbmate.DB
	fs      embed.FS
	connStr string
	l       *slog.Logger
}

// newSQLiteMigrator creates a new SQLite migrator. The connection string should be a file path or ":memory:" for in-memory databases.
func newSQLiteMigrator(l *slog.Logger, fs embed.FS, connStr string) (*sqliteMigrator, error) {
	if connStr == "" {
		return nil, errors.New("connection string is required")
	}

	_, err := fs.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	if strings.Contains(connStr, "memory") {
		return nil, errors.New("in-memory databases are not supported")
	}

	u, err := url.Parse("sqlite:" + connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	db := dbmate.New(u)
	db.Strict = true
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
	db.AutoDumpSchema = false

	l = l.With(slog.String("component", "db-migrator"), slog.String("dialect", "sqlite"))
	db.Log = utils.NewSlogWriter(l)

	return &sqliteMigrator{
		l:       l,
		db:      db,
		fs:      fs,
		connStr: connStr,
	}, nil
}

// Migrate runs migrations on the SQLite database.
func (m *sqliteMigrator) Migrate() error {
	m.l.Info("Migrating database")

	if err := m.db.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// DumpSchema dumps the SQLite database schema to the specified file path.
func (m *sqliteMigrator) DumpSchema(filePath string) error {
	m.db.SchemaFile = filePath

	m.l.Info("Dumping schema", slog.String("file", filePath))

	if err := m.db.DumpSchema(); err != nil {
		return fmt.Errorf("failed to dump schema: %w", err)
	}

	return nil
}
