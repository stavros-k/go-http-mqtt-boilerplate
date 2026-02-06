package dialect

import (
	"embed"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/database/postgres"
	"http-mqtt-boilerplate/backend/internal/database/sqlite"
)

type Dialect string

const (
	SQLite     Dialect = "sqlite"
	PostgreSQL Dialect = "postgres"
)

func (d Dialect) Validate() error {
	switch d {
	case SQLite, PostgreSQL:
		return nil
	default:
		return fmt.Errorf("unsupported dialect: %s", d)
	}
}

func (d Dialect) String() string {
	return string(d)
}

func (d Dialect) Driver() string {
	switch d {
	case SQLite:
		return "sqlite3"
	case PostgreSQL:
		return "pgx"
	default:
		return ""
	}
}

func (d Dialect) MigrationFS() embed.FS {
	switch d {
	case SQLite:
		return sqlite.GetMigrationsFS()
	case PostgreSQL:
		return postgres.GetMigrationsFS()
	default:
		return embed.FS{}
	}
}
