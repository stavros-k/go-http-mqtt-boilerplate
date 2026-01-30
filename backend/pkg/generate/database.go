package generate

import (
	"bytes"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/database/sqlite"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"
	"strings"
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

type DatabaseStats struct {
	TableCount int `json:"tableCount"`
}

func (g *OpenAPICollector) GetDatabaseStats(schema string) (*DatabaseStats, error) {
	g.l.Debug("Getting database stats")

	tableCount := strings.Count(schema, "CREATE TABLE")

	return &DatabaseStats{tableCount}, nil
}
