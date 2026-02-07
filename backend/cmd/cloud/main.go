package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	cloudapi "http-mqtt-boilerplate/backend/internal/cloud/api"
	clouddb "http-mqtt-boilerplate/backend/internal/cloud/gen"
	cloudservices "http-mqtt-boilerplate/backend/internal/cloud/services"
	"http-mqtt-boilerplate/backend/internal/config"
	apicommon "http-mqtt-boilerplate/backend/internal/shared/api"
	sharedapi "http-mqtt-boilerplate/backend/internal/shared/api"
	"http-mqtt-boilerplate/backend/internal/shared/helpers"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"
)

func main() {
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	logger := slog.Default()
	config, err := config.New()
	if err != nil {
		fatalIfErr(logger, fmt.Errorf("failed to create config: %w", err))
	}

	defer func() {
		if err := config.Close(); err != nil {
			logger.Error("failed to close config", utils.ErrAttr(err))
		}
	}()

	// Initialize logger
	logger = helpers.GetLogger(config)

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
		err := helpers.RunMigrations(logger, config, "shared/migrations", "cloud/migrations")
		fatalIfErr(logger, err)

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
	httpServer := apicommon.NewHTTPServer(logger, httpAddr, rb.Router())
	httpServer.StartOnBackground(sigCancel)

	// Wait for signal (either OS or some failure)
	<-sigCtx.Done()
	logger.Info("received signal, shutting down...")

	// Shutdown HTTP server
	logger.Info("http server shutting down...")

	if err := httpServer.ShutdownWithDefaultTimeout(context.Background()); err != nil {
		logger.Error("http server shutdown failed", utils.ErrAttr(err))
	}

	logger.Info("server exited gracefully")
}

// registerHTTPHandlers registers all HTTP handlers.
func registerHTTPHandlers(l *slog.Logger, rb *router.RouteBuilder, h *cloudapi.Handler) {
	l.Info("Registering HTTP handlers...")

	// Create middleware handler
	mw := sharedapi.NewMiddlewareHandler(l)

	rb.Route("/api", func(rb *router.RouteBuilder) {
		// Add recoverer
		rb.Use(mw.RecoveryMiddleware)
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
			"backend/internal/shared/types",    // Shared types (ErrorResponse, PingResponse, etc.)
			"backend/internal/cloud/api/types", // Cloud API types
		},
		DatabaseSchemaFileOutputPath: "docs/cloud/schema.sql",
		DocsFileOutputPath:           "docs/cloud/api_docs.json",
		OpenAPISpecOutputPath:        "docs/cloud/openapi.yaml",
		Deployment:                   "cloud",
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

func fatalIfErr(l *slog.Logger, err error) {
	if err == nil {
		return
	}

	l.Error("error", utils.ErrAttr(err))
	os.Exit(1)
}
