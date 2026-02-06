package main

import (
	"context"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/api"
	"http-mqtt-boilerplate/backend/internal/config"
	postgresgen "http-mqtt-boilerplate/backend/internal/database/postgres/gen"
	"http-mqtt-boilerplate/backend/internal/services"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	shutdownTimeout   = 30 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	config, err := config.New(dialect.PostgreSQL)
	if err != nil {
		fatalIfErr(slog.Default(), fmt.Errorf("failed to create config: %w", err))
	}

	defer utils.LogOnError(slog.Default(), config.Close, "failed to close config")

	// Initialize logger
	logger := getLogger(config)

	// Run migrations before opening database connection
	if err := runMigrations(logger, config); err != nil {
		fatalIfErr(logger, fmt.Errorf("failed to run migrations: %w", err))
	}

	db, err := pgx.Connect(context.TODO(), config.Database)
	fatalIfErr(logger, err)

	defer func() {
		if err := db.Close(context.TODO()); err != nil {
			logger.Error("failed to close database", utils.ErrAttr(err))
		}
	}()

	queries := postgresgen.New(db)
	_ = queries

	// Create collector for OpenAPI generation
	collector, err := getCollector(config, logger)
	fatalIfErr(logger, err)

	// Builders
	rb, err := router.NewRouteBuilder(logger, collector)
	fatalIfErr(logger, err)


	// Now create services with the initialized MQTT client
	services := services.NewServices(logger, nil, nil, nil)
	apiHandler := api.NewAPIHandler(logger, services)

	registerHTTPHandlers(logger, rb, apiHandler)

	if config.Generate {
		if err := collector.Generate(); err != nil {
			fatalIfErr(logger, fmt.Errorf("failed to generate API documentation: %w", err))
		}

		return
	}

	// HTTP Server
	httpAddr := fmt.Sprintf(":%d", config.Port)
	httpServer := &http.Server{
		Addr:              httpAddr,
		Handler:           rb.Router(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		logger.Info("http server listening", slog.String("address", httpAddr))

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", utils.ErrAttr(err))
			sigCancel()
		}
	}()

	// Wait for signal (either OS or some failure)
	<-sigCtx.Done()
	logger.Info("received signal, shutting down...")

	// Shutdown HTTP server
	logger.Info("http server shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", utils.ErrAttr(err))
	}


	logger.Info("server exited gracefully")
}


// registerHTTPHandlers registers all HTTP handlers.
func registerHTTPHandlers(l *slog.Logger, rb *router.RouteBuilder, h *api.Handler) {
	l.Info("Registering HTTP handlers...")
	rb.Route("/api", func(rb *router.RouteBuilder) {
		// Add request ID
		rb.Use(h.RequestIDMiddleware)
		// Add request logger
		rb.Use(h.LoggerMiddleware)

		h.RegisterPing("/ping", rb)
		h.RegisterHealth("/health", rb)

		rb.Route("/team", func(rb *router.RouteBuilder) {
			h.RegisterGetTeam("/{teamID}", rb)
			h.RegisterPutTeam("/", rb)
			h.RegisterCreateTeam("/", rb)
			h.RegisterDeleteTeam("/", rb)
		})
	})

	webapp, err := web.DocsApp()
	fatalIfErr(l, err)
	webapp.Register(rb.Router(), l)

	rb.Router().HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusMovedPermanently)
	})

	l.Info("HTTP handlers registered successfully")
}


//nolint:ireturn // Returns MetadataCollector interface (OpenAPICollector or NoopCollector)
func getCollector(c *config.Config, l *slog.Logger) (generate.MetadataCollector, error) {
	if !c.Generate {
		return &generate.NoopCollector{}, nil
	}

	return generate.NewOpenAPICollector(l, generate.OpenAPICollectorOptions{
		GoTypesDirPath:               "backend/pkg/apitypes",
		DatabaseSchemaFileOutputPath: "api_cloud/schema.sql",
		DocsFileOutputPath:           "api_cloud/api_docs.json",
		OpenAPISpecOutputPath:        "api_cloud/openapi.yaml",
		Dialect:                      c.Dialect,
		APIInfo: generate.APIInfo{
			Title:       "Cloud API",
			Version:     utils.GetVersionShort(),
			Description: "Cloud API Documentation",
			Servers: []generate.ServerInfo{
				{URL: "http://localhost:8080", Description: "Local server"},
			},
		},
	})
}

func getLogger(config *config.Config) *slog.Logger {
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

func fatalIfErr(l *slog.Logger, err error) {
	if err == nil {
		return
	}

	l.Error("error", utils.ErrAttr(err))
	os.Exit(1)
}

func runMigrations(l *slog.Logger, c *config.Config) error {
	l.Info("Running database migrations", slog.String("dialect", c.Dialect.String()))

	// Create migrator
	mig, err := migrator.New(l, c.Dialect, c.Dialect.MigrationFS(), c.Database)
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
