package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"http-mqtt-boilerplate/backend/internal/apicommon"
	"http-mqtt-boilerplate/backend/internal/cloudapi"
	"http-mqtt-boilerplate/backend/internal/config"
	clouddb "http-mqtt-boilerplate/backend/internal/database/clouddb/gen"
	cloudservices "http-mqtt-boilerplate/backend/internal/services/cloud"
	"http-mqtt-boilerplate/backend/pkg/dialect"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"
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

	// Create collector for OpenAPI generation
	collector, err := getCollector(config, logger)
	fatalIfErr(logger, err)

	// Conditionally initialize database and queries
	var (
		pool    *pgxpool.Pool    = nil
		queries *clouddb.Queries = nil
	)

	if !config.Generate {
		// For runtime, initialize database
		if err := runMigrations(logger, config); err != nil {
			fatalIfErr(logger, fmt.Errorf("failed to run migrations: %w", err))
		}

		pool, err = pgxpool.New(context.TODO(), config.Database)
		fatalIfErr(logger, err)

		defer func() {
			pool.Close()
		}()

		queries = clouddb.New(pool)
	}

	// Builders
	rb, err := router.NewRouteBuilder(logger, collector)
	fatalIfErr(logger, err)

	// Create services
	services := cloudservices.NewServices(logger, pool, queries)
	apiHandler := cloudapi.NewHandler(logger, services)

	registerHTTPHandlers(logger, rb, apiHandler)

	// If generating, generate and exit
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
func registerHTTPHandlers(l *slog.Logger, rb *router.RouteBuilder, h *cloudapi.Handler) {
	l.Info("Registering HTTP handlers...")

	// Create middleware handler
	mw := apicommon.NewMiddlewareHandler(l)

	rb.Route("/api", func(rb *router.RouteBuilder) {
		// Add request ID
		rb.Use(mw.RequestIDMiddleware)
		// Add request logger
		rb.Use(mw.LoggerMiddleware)

		h.RegisterPing("/ping", rb)
		h.RegisterHealth("/health", rb)
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
		GoTypesDirPaths: []string{
			"backend/pkg/types/common",   // Common types (ErrorResponse, etc.)
			"backend/pkg/types/cloudapi", // Cloud-specific types
		},
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
