package migrator

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/dbstats"
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

// GetDatabaseStats queries the PostgreSQL database directly to retrieve metadata about tables, columns, foreign keys, and indexes.
func (m *postgresMigrator) GetDatabaseStats() (dbstats.DatabaseStats, error) {
	m.l.Debug("Getting database stats from PostgreSQL")

	// Open a connection to PostgreSQL using the internal connection string (use pgx driver)
	db, err := sql.Open("pgx", m.connStr)
	if err != nil {
		return dbstats.DatabaseStats{}, fmt.Errorf("failed to open database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			m.l.Error("failed to close database", utils.ErrAttr(err))
		}
	}()

	ctx := context.Background()
	schemaName := "public" // Default schema where migrations run

	// Get all table names
	tableNames, err := m.getTableNames(ctx, db, schemaName)
	if err != nil {
		return dbstats.DatabaseStats{}, err
	}

	// Get metadata for each table
	var tables []dbstats.Table

	for _, tableName := range tableNames {
		table, err := m.getTableMetadata(ctx, db, schemaName, tableName)
		if err != nil {
			return dbstats.DatabaseStats{}, err
		}

		tables = append(tables, table)
	}

	return dbstats.DatabaseStats{Tables: tables}, nil
}

func (m *postgresMigrator) getTableNames(ctx context.Context, db *sql.DB, schemaName string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = $1
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	defer func() {
		if err := rows.Err(); err != nil {
			m.l.Error("failed to iterate tables", utils.ErrAttr(err))
		}
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
		return nil, fmt.Errorf("error iterating table names: %w", err)
	}

	return tableNames, nil
}

func (m *postgresMigrator) getTableMetadata(ctx context.Context, db *sql.DB, schemaName, tableName string) (dbstats.Table, error) {
	table := dbstats.Table{Name: tableName}

	columns, err := m.getTableColumns(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}

	table.Columns = columns

	foreignKeys, err := m.getTableForeignKeys(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}

	table.ForeignKeys = foreignKeys

	indexes, err := m.getTableIndexes(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}

	table.Indexes = indexes

	return table, nil
}

func (m *postgresMigrator) getTableColumns(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]dbstats.Column, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			CASE WHEN pk.column_name IS NOT NULL THEN true ELSE false END as is_primary
		FROM information_schema.columns c
		LEFT JOIN (
			SELECT kcu.column_name
			FROM information_schema.table_constraints tc
			JOIN information_schema.key_column_usage kcu
				ON tc.constraint_name = kcu.constraint_name
				AND tc.table_schema = kcu.table_schema
			WHERE tc.constraint_type = 'PRIMARY KEY'
				AND tc.table_name = $1
				AND tc.table_schema = $2
		) pk ON c.column_name = pk.column_name
		WHERE c.table_name = $1
		  AND c.table_schema = $2
		ORDER BY c.ordinal_position
	`, tableName, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for %s: %w", tableName, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			m.l.Error("failed to close columns rows", utils.ErrAttr(err))
		}
	}()

	var columns []dbstats.Column

	for rows.Next() {
		var (
			c          dbstats.Column
			isNullable string
			dflt       sql.NullString
		)

		if err := rows.Scan(&c.Name, &c.Type, &isNullable, &dflt, &c.PrimaryKey); err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		c.NotNull = isNullable == "NO" || c.PrimaryKey // PRIMARY KEY implies NOT NULL
		if dflt.Valid {
			c.Default = &dflt.String
		}

		columns = append(columns, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating columns for %s: %w", tableName, err)
	}

	return columns, nil
}

func (m *postgresMigrator) getTableForeignKeys(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]dbstats.ForeignKey, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			kcu.column_name as from_column,
			ccu.table_name AS foreign_table_name,
			ccu.column_name AS foreign_column_name
		FROM information_schema.table_constraints AS tc
		JOIN information_schema.key_column_usage AS kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage AS ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_name = $1
			AND tc.table_schema = $2
	`, tableName, schemaName)
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys for %s: %w", tableName, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			m.l.Error("failed to close foreign keys rows", utils.ErrAttr(err))
		}
	}()

	var foreignKeys []dbstats.ForeignKey

	for rows.Next() {
		var fk dbstats.ForeignKey
		if err := rows.Scan(&fk.From, &fk.Table, &fk.To); err != nil {
			return nil, fmt.Errorf("failed to scan foreign key: %w", err)
		}

		foreignKeys = append(foreignKeys, fk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating foreign keys for %s: %w", tableName, err)
	}

	return foreignKeys, nil
}

func (m *postgresMigrator) getTableIndexes(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]dbstats.Index, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			i.indexname,
			ix.indisunique,
			array_agg(a.attname ORDER BY array_position(ix.indkey::int[], a.attnum))
		FROM pg_indexes i
		JOIN pg_class c ON c.relname = i.tablename
		JOIN pg_index ix ON ix.indexrelid = (
			SELECT oid FROM pg_class WHERE relname = i.indexname AND relnamespace = c.relnamespace
		)
		JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY(ix.indkey)
		WHERE i.schemaname = $1
			AND i.tablename = $2
			AND NOT ix.indisprimary
		GROUP BY i.indexname, ix.indisunique
	`, schemaName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes for %s: %w", tableName, err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			m.l.Error("failed to close indexes rows", utils.ErrAttr(err))
		}
	}()

	var indexes []dbstats.Index

	for rows.Next() {
		var (
			idx     dbstats.Index
			colsStr string
		)

		if err := rows.Scan(&idx.Name, &idx.Unique, &colsStr); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		idx.Columns = parsePostgresArray(colsStr)
		indexes = append(indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating indexes for %s: %w", tableName, err)
	}

	return indexes, nil
}

// parsePostgresArray parses a PostgreSQL array string like "{col1,col2,col3}" into a Go slice.
func parsePostgresArray(s string) []string {
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return nil
	}

	// Remove braces and split by comma
	inner := s[1 : len(s)-1]
	if inner == "" {
		return []string{}
	}

	// Simple split - doesn't handle quoted elements with commas, but column names don't have commas
	parts := make([]string, 0, 10)
	parts = append(parts, splitPostgresArray(inner)...)

	return parts
}

// splitPostgresArray splits a PostgreSQL array content by commas.
func splitPostgresArray(s string) []string {
	if s == "" {
		return []string{}
	}

	var (
		result  []string
		current string
	)

	inQuotes := false

	for i := range len(s) {
		ch := s[i]

		if ch == '"' {
			inQuotes = !inQuotes

			continue
		}

		if ch == ',' && !inQuotes {
			result = append(result, current)
			current = ""

			continue
		}

		current += string(ch)
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
