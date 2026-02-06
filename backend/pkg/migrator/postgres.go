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
	_ "github.com/amacneil/dbmate/v2/pkg/driver/postgres"
	_ "github.com/lib/pq"
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

	// Parse the connection string to build the postgres:// URL for dbmate
	var u *url.URL
	if strings.HasPrefix(connStr, "postgres://") || strings.HasPrefix(connStr, "postgresql://") {
		u, err = url.Parse(connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection string: %w", err)
		}
	} else {
		// Convert key=value format to URL format
		u, err = parsePostgresConnString(connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse connection string: %w", err)
		}
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

	return nil
}

// GetDatabaseStats connects to the PostgreSQL database and retrieves metadata about tables, columns, foreign keys, and indexes.
func (m *postgresMigrator) GetDatabaseStats() (dbstats.DatabaseStats, error) {
	m.l.Debug("Getting database stats")

	ctx := context.Background()

	// Open a connection to PostgreSQL using the internal connection string
	// This connects to the same database that was migrated
	db, err := sql.Open("postgres", m.connStr)
	if err != nil {
		return dbstats.DatabaseStats{}, fmt.Errorf("failed to open database: %w", err)
	}
	defer utils.LogOnError(m.l, db.Close, "failed to close database")

	// Query the public schema (default schema where migrations ran)
	schemaName := "public"

	tableNames, err := m.getTableNames(ctx, db, schemaName)
	if err != nil {
		return dbstats.DatabaseStats{}, err
	}

	var tables []dbstats.Table
	for _, name := range tableNames {
		table, err := m.getTableMetadata(ctx, db, schemaName, name)
		if err != nil {
			return dbstats.DatabaseStats{}, err
		}
		tables = append(tables, table)
	}

	return dbstats.DatabaseStats{Tables: tables}, nil
}

// getTableNames retrieves all table names from the PostgreSQL schema.
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
func (m *postgresMigrator) getTableMetadata(ctx context.Context, db *sql.DB, schemaName, tableName string) (dbstats.Table, error) {
	t := dbstats.Table{Name: tableName}

	columns, err := m.getTableColumns(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}
	t.Columns = columns

	foreignKeys, err := m.getTableForeignKeys(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}
	t.ForeignKeys = foreignKeys

	indexes, err := m.getTableIndexes(ctx, db, schemaName, tableName)
	if err != nil {
		return dbstats.Table{}, err
	}
	t.Indexes = indexes

	return t, nil
}

// getTableColumns retrieves all columns for a specific table.
func (m *postgresMigrator) getTableColumns(ctx context.Context, db *sql.DB, schemaName, tableName string) ([]dbstats.Column, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			column_name,
			data_type,
			is_nullable,
			column_default,
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
		utils.LogOnError(m.l, rows.Close, "failed to close column rows for table "+tableName)
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
		return nil, fmt.Errorf("failed to iterate columns for %s: %w", tableName, err)
	}

	return columns, nil
}

// getTableForeignKeys retrieves all foreign keys for a specific table.
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
		utils.LogOnError(m.l, rows.Close, "failed to close foreign key rows for table "+tableName)
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
		return nil, fmt.Errorf("failed to iterate foreign keys for %s: %w", tableName, err)
	}

	return foreignKeys, nil
}

// getTableIndexes retrieves all indexes for a specific table.
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
		utils.LogOnError(m.l, rows.Close, "failed to close index rows for table "+tableName)
	}()

	var indexes []dbstats.Index
	for rows.Next() {
		var (
			idx  dbstats.Index
			cols []string
		)

		if err := rows.Scan(&idx.Name, &idx.Unique, &cols); err != nil {
			return nil, fmt.Errorf("failed to scan index: %w", err)
		}

		idx.Columns = cols
		indexes = append(indexes, idx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate indexes for %s: %w", tableName, err)
	}

	return indexes, nil
}

// parsePostgresConnString converts a key=value connection string to a postgres:// URL.
func parsePostgresConnString(connStr string) (*url.URL, error) {
	params := make(map[string]string)

	// Parse key=value pairs
	pairs := strings.SplitSeq(connStr, " ")
	for pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			params[parts[0]] = parts[1]
		}
	}

	// Build URL
	host := params["host"]
	if host == "" {
		host = "localhost"
	}

	port := params["port"]
	if port == "" {
		port = "5432"
	}

	dbname := params["dbname"]
	if dbname == "" {
		return nil, errors.New("dbname is required in connection string")
	}

	user := params["user"]
	if user == "" {
		user = "postgres"
	}

	password := params["password"]

	sslmode := params["sslmode"]
	if sslmode == "" {
		sslmode = "disable"
	}

	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   "/" + dbname,
	}

	q := u.Query()
	q.Set("sslmode", sslmode)
	u.RawQuery = q.Encode()

	return u, nil
}
