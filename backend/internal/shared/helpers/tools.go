package helpers

import (
	"fmt"
	"http-mqtt-boilerplate/backend/internal/config"
	"http-mqtt-boilerplate/backend/internal/migrations"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
)

func GetLogger(config *config.Config) *slog.Logger {
	logOptions := slog.HandlerOptions{
		Level:       config.LogLevel,
		ReplaceAttr: utils.SlogReplacer,
	}

	var logHandler slog.Handler = slog.NewJSONHandler(config.LogOutput, &logOptions)
	if config.Generate {
		logHandler = slog.NewTextHandler(config.LogOutput, &logOptions)
	}

	return slog.New(logHandler).With(slog.String("version", utils.GetVersionShort()))
}

func RunMigrations(l *slog.Logger, c *config.Config, dirs ...string) error {
	l.Info("Running database migrations")

	// Create migrator with shared + cloud migration directories
	mig, err := migrator.New(l, c.Database, migrations.GetFS(), dirs...)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	// Run migrations
	if err := mig.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}

	l.Info("Database migrations completed successfully")

	return nil
}
