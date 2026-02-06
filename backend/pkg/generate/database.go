package generate

import (
	"bytes"
	"context"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/migrations"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"

	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
)

// GenerateDatabaseSchema runs migrations on a temporary database and returns the resulting schema.
// This generates a SQL schema dump from the application's migrations.
func (g *OpenAPICollector) GenerateDatabaseSchema(deployment string, schemaOutputPath string) (string, error) {
	g.l.Debug("Generating database schema from migrations", slog.String("deployment", deployment))

	// Start a PostgreSQL container for schema generation
	ctx := context.Background()

	container, err := postgrescontainer.Run(ctx,
		"postgres:18-alpine",
		postgrescontainer.WithDatabase("testdb"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpassword"),
		postgrescontainer.BasicWaitStrategies(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	defer func() {
		if err := container.Terminate(ctx); err != nil {
			g.l.Error("failed to terminate PostgreSQL container", utils.ErrAttr(err))
		}
	}()

	// Get connection string from container
	tempDB, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("failed to get connection string: %w", err)
	}

	// Get migrations based on deployment
	var mig migrator.Migrator
	var migrationDirs []string
	switch deployment {
	case "local":
		migrationDirs = []string{"shared/migrations", "local/migrations"}
	case "cloud":
		migrationDirs = []string{"shared/migrations", "cloud/migrations"}
	default:
		return "", fmt.Errorf("unsupported deployment: %s", deployment)
	}

	mig, err = migrator.New(g.l, tempDB, migrations.GetFS(), migrationDirs...)
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

	g.l.Info("Database schema generated", slog.String("file", schemaOutputPath), slog.String("deployment", deployment))

	return schema, nil
}
