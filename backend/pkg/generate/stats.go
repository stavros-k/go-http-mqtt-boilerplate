package generate

import (
	"context"
	"database/sql"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

type DatabaseStats struct {
	Tables []Table `json:"tables"`
}

func (d *DatabaseStats) NonNil() {
	if d.Tables == nil {
		d.Tables = []Table{}
	}

	for i := range d.Tables {
		d.Tables[i].NonNil()
	}
}

type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreignKeys"`
	Indexes     []Index      `json:"indexes"`
}

func (t *Table) NonNil() {
	if t.Columns == nil {
		t.Columns = []Column{}
	}

	if t.ForeignKeys == nil {
		t.ForeignKeys = []ForeignKey{}
	}

	if t.Indexes == nil {
		t.Indexes = []Index{}
	}
}

type Column struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	NotNull    bool    `json:"notNull"`
	Default    *string `json:"default,omitempty"`
	PrimaryKey bool    `json:"primaryKey"`
}

type ForeignKey struct {
	From  string `json:"from"`
	Table string `json:"table"`
	To    string `json:"to"`
}

type Index struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
}

// GetDatabaseStats retrieves database metadata based on the dialect and connection string.
func GetDatabaseStats(l *slog.Logger, d dialect.Dialect, connStr string) (DatabaseStats, error) {
	switch d {
	case dialect.SQLite:
		return getSQLiteStats(l, connStr)
	case dialect.PostgreSQL:
		return getPostgresStats(l, connStr)
	default:
		return DatabaseStats{}, fmt.Errorf("unsupported dialect: %s", d)
	}
}

// getSQLiteStats connects to the SQLite database and retrieves metadata about tables, columns, foreign keys, and indexes.
func getSQLiteStats(l *slog.Logger, connStr string) (DatabaseStats, error) {
	l.Debug("Getting database stats from SQLite")

	ctx := context.Background()

	// Open a new connection to the SQLite database
	db, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return DatabaseStats{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer utils.LogOnError(l, db.Close, "failed to close database")

	tableNames, err := getSQLiteTableNames(ctx, l, db)
	if err != nil {
		return DatabaseStats{}, err
	}

	var tables []Table

	for _, name := range tableNames {
		table, err := getSQLiteTableMetadata(ctx, l, db, name)
		if err != nil {
			return DatabaseStats{}, err
		}

		tables = append(tables, table)
	}

	return DatabaseStats{Tables: tables}, nil
}

// getSQLiteTableNames retrieves all table names from the SQLite database.
func getSQLiteTableNames(ctx context.Context, l *slog.Logger, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}

	defer func() {
		utils.LogOnError(l, rows.Close, "failed to close rows for tables query")
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

// getSQLiteTableMetadata retrieves all metadata for a specific table.
func getSQLiteTableMetadata(ctx context.Context, l *slog.Logger, db *sql.DB, name string) (Table, error) {
	t := Table{Name: name}

	columns, err := getSQLiteTableColumns(ctx, l, db, name)
	if err != nil {
		return Table{}, err
	}

	t.Columns = columns

	foreignKeys, err := getSQLiteTableForeignKeys(ctx, l, db, name)
	if err != nil {
		return Table{}, err
	}

	t.ForeignKeys = foreignKeys

	indexes, err := getSQLiteTableIndexes(ctx, l, db, name)
	if err != nil {
		return Table{}, err
	}

	t.Indexes = indexes

	return t, nil
}

// getSQLiteTableColumns retrieves all columns for a specific table.
func getSQLiteTableColumns(ctx context.Context, l *slog.Logger, db *sql.DB, tableName string) ([]Column, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query columns for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(l, rows.Close, "failed to close column rows for table "+tableName)
	}()

	var columns []Column

	for rows.Next() {
		var (
			c                Column
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

// getSQLiteTableForeignKeys retrieves all foreign keys for a specific table.
func getSQLiteTableForeignKeys(ctx context.Context, l *slog.Logger, db *sql.DB, tableName string) ([]ForeignKey, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query foreign keys for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(l, rows.Close, "failed to close foreign key rows for table "+tableName)
	}()

	var foreignKeys []ForeignKey

	for rows.Next() {
		var (
			id, seq                   int
			fk                        ForeignKey
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

// getSQLiteTableIndexes retrieves all indexes for a specific table.
func getSQLiteTableIndexes(ctx context.Context, l *slog.Logger, db *sql.DB, tableName string) ([]Index, error) {
	type indexMeta struct {
		Name   string
		Unique bool
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_list("%s")`, tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to query indexes for %s: %w", tableName, err)
	}

	defer func() {
		utils.LogOnError(l, rows.Close, "failed to close index rows for table "+tableName)
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

	var indexes []Index

	for _, meta := range indexMetas {
		idx, err := getSQLiteIndexColumns(ctx, l, db, meta.Name, meta.Unique)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, idx)
	}

	return indexes, nil
}

// getSQLiteIndexColumns retrieves all columns for a specific index.
func getSQLiteIndexColumns(ctx context.Context, l *slog.Logger, db *sql.DB, indexName string, unique bool) (Index, error) {
	idx := Index{Name: indexName, Unique: unique}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_info("%s")`, indexName))
	if err != nil {
		return Index{}, fmt.Errorf("failed to query index info for %s: %w", indexName, err)
	}

	defer func() {
		utils.LogOnError(l, rows.Close, "failed to close index info rows for index "+indexName)
	}()

	for rows.Next() {
		var (
			seqno, cid int
			colName    string
		)

		if err := rows.Scan(&seqno, &cid, &colName); err != nil {
			return Index{}, fmt.Errorf("failed to scan index info: %w", err)
		}

		idx.Columns = append(idx.Columns, colName)
	}

	if err := rows.Err(); err != nil {
		return Index{}, fmt.Errorf("failed to iterate index info for %s: %w", indexName, err)
	}

	return idx, nil
}

// getPostgresStats queries the PostgreSQL database directly to retrieve metadata about tables, columns, foreign keys, and indexes.
func getPostgresStats(l *slog.Logger, connStr string) (DatabaseStats, error) {
	l.Debug("Getting database stats from PostgreSQL")

	// Open a connection to PostgreSQL using the internal connection string (use pgx driver)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return DatabaseStats{}, fmt.Errorf("failed to open database: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			l.Error("failed to close database", utils.ErrAttr(err))
		}
	}()

	ctx := context.Background()
	schemaName := "public" // Default schema where migrations run

	// Get all table names
	tableNames, err := getPostgresTableNames(ctx, l, db, schemaName)
	if err != nil {
		return DatabaseStats{}, err
	}

	// Get metadata for each table
	var tables []Table

	for _, tableName := range tableNames {
		table, err := getPostgresTableMetadata(ctx, l, db, schemaName, tableName)
		if err != nil {
			return DatabaseStats{}, err
		}

		tables = append(tables, table)
	}

	return DatabaseStats{Tables: tables}, nil
}

func getPostgresTableNames(ctx context.Context, l *slog.Logger, db *sql.DB, schemaName string) ([]string, error) {
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
		if err := rows.Close(); err != nil {
			l.Error("failed to close tables rows", utils.ErrAttr(err))
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

func getPostgresTableMetadata(ctx context.Context, l *slog.Logger, db *sql.DB, schemaName, tableName string) (Table, error) {
	table := Table{Name: tableName}

	columns, err := getPostgresTableColumns(ctx, l, db, schemaName, tableName)
	if err != nil {
		return Table{}, err
	}

	table.Columns = columns

	foreignKeys, err := getPostgresTableForeignKeys(ctx, l, db, schemaName, tableName)
	if err != nil {
		return Table{}, err
	}

	table.ForeignKeys = foreignKeys

	indexes, err := getPostgresTableIndexes(ctx, l, db, schemaName, tableName)
	if err != nil {
		return Table{}, err
	}

	table.Indexes = indexes

	return table, nil
}

func getPostgresTableColumns(ctx context.Context, l *slog.Logger, db *sql.DB, schemaName, tableName string) ([]Column, error) {
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
			l.Error("failed to close columns rows", utils.ErrAttr(err))
		}
	}()

	var columns []Column

	for rows.Next() {
		var (
			c          Column
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

func getPostgresTableForeignKeys(ctx context.Context, l *slog.Logger, db *sql.DB, schemaName, tableName string) ([]ForeignKey, error) {
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
			l.Error("failed to close foreign keys rows", utils.ErrAttr(err))
		}
	}()

	var foreignKeys []ForeignKey

	for rows.Next() {
		var fk ForeignKey
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

func getPostgresTableIndexes(ctx context.Context, l *slog.Logger, db *sql.DB, schemaName, tableName string) ([]Index, error) {
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
			l.Error("failed to close indexes rows", utils.ErrAttr(err))
		}
	}()

	var indexes []Index

	for rows.Next() {
		var (
			idx     Index
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
