package generate

import (
	"bytes"
	"database/sql"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/database/sqlite"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"
)

// GenerateDatabaseSchema runs migrations on a temporary database and returns the resulting schema.
// This generates a SQL schema dump from the application's migrations.
func (g *OpenAPICollector) GenerateDatabaseSchema(schemaOutputPath string) (string, error) {
	g.l.Debug("Generating database schema from migrations")

	// Create a temporary database file
	tempDBFile, err := os.CreateTemp(os.TempDir(), "temp-db-*.sqlite")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary database file: %w", err)
	}

	// Close immediately - we only need the file path, not the handle
	if err := tempDBFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary database file: %w", err)
	}

	defer func() {
		if err := os.Remove(tempDBFile.Name()); err != nil {
			g.l.Error("failed to remove temporary database file", utils.ErrAttr(err))
		}
	}()

	// Create a migrator for the temporary database
	mig, err := migrator.New(g.l, sqlite.GetMigrationsFS(), tempDBFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := mig.Migrate(); err != nil {
		return "", fmt.Errorf("failed to migrate database: %w", err)
	}

	// Dump the database schema to the specified output path
	if err = mig.DumpSchema(schemaOutputPath); err != nil {
		return "", fmt.Errorf("failed to dump schema: %w", err)
	}

	// Read the schema file
	schemaBytes, err := os.ReadFile(schemaOutputPath)
	if err != nil {
		return "", fmt.Errorf("failed to read schema file: %w", err)
	}

	g.l.Info("Database schema generated", slog.String("file", schemaOutputPath))

	return string(bytes.TrimSpace(schemaBytes)), nil
}

func (g *OpenAPICollector) GetDatabaseStats(schema string) (DatabaseStats, error) {
	g.l.Debug("Getting database stats")

	// Open an in-memory SQLite database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return DatabaseStats{}, fmt.Errorf("failed to open in-memory database: %w", err)
	}
	defer db.Close()

	// Apply the schema
	if _, err := db.Exec(schema); err != nil {
		return DatabaseStats{}, fmt.Errorf("failed to apply schema: %w", err)
	}

	// Get all tables
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return DatabaseStats{}, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []Table
	var tableNames []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return DatabaseStats{}, fmt.Errorf("failed to scan table name: %w", err)
		}
		tableNames = append(tableNames, name)
	}

	for _, name := range tableNames {
		t := Table{Name: name}

		// Get columns - collect metadata first to avoid nested query issues
		colRows, err := db.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, name))
		if err != nil {
			return DatabaseStats{}, fmt.Errorf("failed to query columns for %s: %w", name, err)
		}
		for colRows.Next() {
			var c Column
			var cid, notnull, pk int
			var dflt sql.NullString
			if err := colRows.Scan(&cid, &c.Name, &c.Type, &notnull, &dflt, &pk); err != nil {
				return DatabaseStats{}, fmt.Errorf("failed to scan column: %w", err)
			}
			c.NotNull = notnull == 1
			c.PrimaryKey = pk > 0
			if dflt.Valid {
				c.Default = &dflt.String
			}
			t.Columns = append(t.Columns, c)
		}
		colRows.Close()

		// Get foreign keys
		fkRows, err := db.Query(fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, name))
		if err != nil {
			return DatabaseStats{}, fmt.Errorf("failed to query foreign keys for %s: %w", name, err)
		}
		for fkRows.Next() {
			var id, seq int
			var fk ForeignKey
			var onUpdate, onDelete, match string
			if err := fkRows.Scan(&id, &seq, &fk.Table, &fk.From, &fk.To, &onUpdate, &onDelete, &match); err != nil {
				return DatabaseStats{}, fmt.Errorf("failed to scan foreign key: %w", err)
			}
			t.ForeignKeys = append(t.ForeignKeys, fk)
		}
		fkRows.Close()

		// Indexes - collect metadata first to avoid nested query issues
		type indexMeta struct {
			Name   string
			Unique bool
		}
		var indexes []indexMeta

		idxRows, err := db.Query(fmt.Sprintf(`PRAGMA index_list("%s")`, name))
		if err != nil {
			return DatabaseStats{}, fmt.Errorf("failed to query indexes for %s: %w", name, err)
		}
		for idxRows.Next() {
			var seq, unique int
			var idxName, origin, partial string
			if err := idxRows.Scan(&seq, &idxName, &unique, &origin, &partial); err != nil {
				return DatabaseStats{}, fmt.Errorf("failed to scan index: %w", err)
			}
			if origin == "pk" {
				continue
			}
			indexes = append(indexes, indexMeta{Name: idxName, Unique: unique == 1})
		}
		idxRows.Close()

		// Now query each index's columns separately
		for _, meta := range indexes {
			idx := Index{Name: meta.Name, Unique: meta.Unique}

			infoRows, err := db.Query(fmt.Sprintf(`PRAGMA index_info("%s")`, meta.Name))
			if err != nil {
				return DatabaseStats{}, fmt.Errorf("failed to query index info for %s: %w", meta.Name, err)
			}
			for infoRows.Next() {
				var seqno, cid int
				var colName string
				if err := infoRows.Scan(&seqno, &cid, &colName); err != nil {
					return DatabaseStats{}, fmt.Errorf("failed to scan index info: %w", err)
				}
				idx.Columns = append(idx.Columns, colName)
			}
			infoRows.Close()

			t.Indexes = append(t.Indexes, idx)
		}

		tables = append(tables, t)
	}

	return DatabaseStats{Tables: tables}, nil
}
