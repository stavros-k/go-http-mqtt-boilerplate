package generate

import (
	"bytes"
	"context"
	"fmt"
	postgres "http-mqtt-boilerplate/backend/internal/database/clouddb"
	sqlite "http-mqtt-boilerplate/backend/internal/database/localdb"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"

	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// GenerateDatabaseSchema runs migrations on a temporary database and returns the resulting schema.
// This generates a SQL schema dump from the application's migrations.
func (g *OpenAPICollector) GenerateDatabaseSchema(d dialect.Dialect, schemaOutputPath string) (string, error) {
	g.l.Debug("Generating database schema from migrations", slog.String("dialect", d.String()))

	// Create temporary database based on dialect
	var (
		tempDB  string
		cleanup func()
	)

	switch d {
	case dialect.SQLite:
		// Create a temporary database file for SQLite
		// We can't use :memory: because each new connection creates a fresh empty database
		tempDBFile, err := os.CreateTemp(os.TempDir(), "temp-db-*.sqlite")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary database file: %w", err)
		}

		// Close immediately - we only need the file path, not the handle
		if err := tempDBFile.Close(); err != nil {
			if removeErr := os.Remove(tempDBFile.Name()); removeErr != nil {
				g.l.Error("failed to remove temporary database file", slog.String("file", tempDBFile.Name()), utils.ErrAttr(removeErr))
			}

			return "", fmt.Errorf("failed to close temporary database file: %w", err)
		}

		tempDB = tempDBFile.Name()
		cleanup = func() {
			if err := os.Remove(tempDBFile.Name()); err != nil {
				g.l.Error("failed to remove temporary database file", slog.String("file", tempDBFile.Name()), utils.ErrAttr(err))
			}
		}

	case dialect.PostgreSQL:
		// Start a PostgreSQL container for schema generation
		ctx := context.Background()

		container, err := postgrescontainer.Run(ctx,
			"postgres:17-alpine",
			postgrescontainer.WithDatabase("testdb"),
			postgrescontainer.WithUsername("postgres"),
			postgrescontainer.WithPassword("postgres"),
			postgrescontainer.BasicWaitStrategies(),
		)
		if err != nil {
			return "", fmt.Errorf("failed to start PostgreSQL container: %w", err)
		}

		cleanup = func() {
			if err := container.Terminate(ctx); err != nil {
				g.l.Error("failed to terminate PostgreSQL container", utils.ErrAttr(err))
			}
		}
		// Get connection string from container
		tempDB, err = container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			cleanup()

			return "", fmt.Errorf("failed to get connection string: %w", err)
		}

	default:
		return "", fmt.Errorf("unsupported dialect: %s", d)
	}

	defer cleanup()

	// Get migrations FS based on dialect
	var migrationsFS = sqlite.GetMigrationsFS()
	if d == dialect.PostgreSQL {
		migrationsFS = postgres.GetMigrationsFS()
	}

	// Create migrator
	mig, err := migrator.New(g.l, d, migrationsFS, tempDB)
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

	schema := string(bytes.TrimSpace(schemaBytes))

	g.l.Info("Database schema generated", slog.String("file", schemaOutputPath), slog.String("dialect", d.String()))

	return schema, nil
}
