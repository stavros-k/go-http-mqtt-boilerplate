package migrator

import (
	"embed"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

type Migrator struct {
	db      *dbmate.DB
	fs      embed.FS // This must contain a migrations directory
	sqlPath string
	l       *slog.Logger
}

// New creates a new Migrator instance.
// TODO: Set a common set of PRAGMA settings for SQLite connections
// TODO: Test if we can edit db from a db browser while working.
func New(l *slog.Logger, fs embed.FS, sqlPath string) (*Migrator, error) {
	if sqlPath == "" {
		return nil, errors.New("sqlPath is required")
	}

	_, err := fs.ReadDir("migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	u, err := url.Parse("sqlite:" + sqlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database url: %w", err)
	}

	db := dbmate.New(u)
	db.Strict = true
	db.FS = fs
	db.MigrationsDir = []string{"migrations"}
	db.AutoDumpSchema = false

	l = l.With(slog.String("component", "db-migrator"))
	db.Log = utils.NewSlogWriter(l)

	return &Migrator{
		l:       l,
		db:      db,
		fs:      fs,
		sqlPath: sqlPath,
	}, nil
}

func (m *Migrator) Migrate() error {
	m.l.Info("Migrating database")

	if err := m.db.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

func (m *Migrator) DumpSchema(filePath string) error {
	m.db.SchemaFile = filePath

	m.l.Info("Dumping schema", slog.String("file", filePath))

	if err := m.db.DumpSchema(); err != nil {
		return fmt.Errorf("failed to dump schema: %w", err)
	}

	return nil
}
