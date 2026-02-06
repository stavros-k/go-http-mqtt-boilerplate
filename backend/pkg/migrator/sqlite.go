package migrator

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/dbstats"
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

// GetDatabaseStats connects to the SQLite database and retrieves metadata about tables, columns, foreign keys, and indexes.
func (m *sqliteMigrator) GetDatabaseStats() (dbstats.DatabaseStats, error) {
	m.l.Debug("Getting database stats")

	ctx := context.Background()

	// Open a new connection to the SQLite database
	// Note: For file-based databases, this works fine. For :memory: databases,
	// this would create a fresh empty database, so callers should use temp files
	// instead of :memory: when they need to query stats after migration.
	db, err := sql.Open("sqlite3", m.connStr)
	if err != nil {
		return dbstats.DatabaseStats{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer utils.LogOnError(m.l, db.Close, "failed to close database")

	tableNames, err := m.getTableNames(ctx, db)
	if err != nil {
		return dbstats.DatabaseStats{}, err
	}

	var tables []dbstats.Table

	for _, name := range tableNames {
		table, err := m.getTableMetadata(ctx, db, name)
		if err != nil {
			return dbstats.DatabaseStats{}, err
		}

		tables = append(tables, table)
	}

	return dbstats.DatabaseStats{Tables: tables}, nil
}

// getTableNames retrieves all table names from the SQLite database.
func (m *sqliteMigrator) getTableNames(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	defer func() {
		utils.LogOnError(m.l, rows.Close, "failed to close rows for tables query")
	}()

	var tableNames []string

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}

		tableNames = append(tableNames, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tables: %w", err)
	}

	return tableNames, nil
}

// getTableMetadata retrieves all metadata for a specific table.
func (m *sqliteMigrator) getTableMetadata(ctx context.Context, db *sql.DB, name string) (dbstats.Table, error) {
	t := dbstats.Table{Name: name}

	columns, err := m.getTableColumns(ctx, db, name)
	if err != nil {
		return dbstats.Table{}, err
	}

	t.Columns = columns

	foreignKeys, err := m.getTableForeignKeys(ctx, db, name)
	if err != nil {
		return dbstats.Table{}, err
	}

	t.ForeignKeys = foreignKeys

	indexes, err := m.getTableIndexes(ctx, db, name)
	if err != nil {
		return dbstats.Table{}, err
	}

	t.Indexes = indexes

	return t, nil
}

// getTableColumns retrieves all columns for a specific table.
func (m *sqliteMigrator) getTableColumns(ctx context.Context, db *sql.DB, tableName string) ([]dbstats.Column, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(m.l, rows.Close, "failed to close column rows for table "+tableName)
	}()

	var columns []dbstats.Column

	for rows.Next() {
		var (
			c                dbstats.Column
			cid, notnull, pk int
			dflt             sql.NullString
		)

		if err := rows.Scan(&cid, &c.Name, &c.Type, &notnull, &dflt, &pk); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.PrimaryKey = pk > 0

		c.NotNull = notnull == 1 || c.PrimaryKey // PRIMARY KEY implies NOT NULL
		if dflt.Valid {
			c.Default = &dflt.String
		}

		columns = append(columns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate columns for %s: %w", tableName, err)
	}

	return columns, nil
}

// getTableForeignKeys retrieves all foreign keys for a specific table.
func (m *sqliteMigrator) getTableForeignKeys(ctx context.Context, db *sql.DB, tableName string) ([]dbstats.ForeignKey, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(m.l, rows.Close, "failed to close foreign key rows for table "+tableName)
	}()

	var foreignKeys []dbstats.ForeignKey

	for rows.Next() {
		var (
			id, seq                   int
			fk                        dbstats.ForeignKey
			onUpdate, onDelete, match string
		)

		if err := rows.Scan(&id, &seq, &fk.Table, &fk.From, &fk.To, &onUpdate, &onDelete, &match); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		foreignKeys = append(foreignKeys, fk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate foreign keys for %s: %w", tableName, err)
	}

	return foreignKeys, nil
}

// getTableIndexes retrieves all indexes for a specific table.
func (m *sqliteMigrator) getTableIndexes(ctx context.Context, db *sql.DB, tableName string) ([]dbstats.Index, error) {
	type indexMeta struct {
		Name   string
		Unique bool
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_list("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(m.l, rows.Close, "failed to close index rows for table "+tableName)
	}()

	var indexMetas []indexMeta

	for rows.Next() {
		var (
			seq, unique              int
			idxName, origin, partial string
		)

		if err := rows.Scan(&seq, &idxName, &unique, &origin, &partial); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		if origin == "pk" {
			continue
		}

		indexMetas = append(indexMetas, indexMeta{Name: idxName, Unique: unique == 1})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate indexes for %s: %w", tableName, err)
	}

	var indexes []dbstats.Index

	for _, meta := range indexMetas {
		idx, err := m.getIndexColumns(ctx, db, meta.Name, meta.Unique)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// getIndexColumns retrieves all columns for a specific index.
func (m *sqliteMigrator) getIndexColumns(ctx context.Context, db *sql.DB, indexName string, unique bool) (dbstats.Index, error) {
	idx := dbstats.Index{Name: indexName, Unique: unique}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_info("%s")`, indexName))
	if err != nil {
		return dbstats.Index{}, fmt.Errorf("failed to query index info for %s: %w", indexName, err)
	}

	defer func() {
		utils.LogOnError(m.l, rows.Close, "failed to close index info rows for index "+indexName)
	}()

	for rows.Next() {
		var (
			seqno, cid int
			colName    string
		)

		if err := rows.Scan(&seqno, &cid, &colName); err != nil {
			return dbstats.Index{}, fmt.Errorf("failed to scan index info: %w", err)
		}

		idx.Columns = append(idx.Columns, colName)
	}

	if err := rows.Err(); err != nil {
		return dbstats.Index{}, fmt.Errorf("failed to iterate index info for %s: %w", indexName, err)
	}

	return idx, nil
}
