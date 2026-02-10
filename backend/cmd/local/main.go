package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"http-mqtt-boilerplate/backend/internal/config"
	localapi "http-mqtt-boilerplate/backend/internal/local/api"
	localdb "http-mqtt-boilerplate/backend/internal/local/gen"
	mqttapi "http-mqtt-boilerplate/backend/internal/local/mqtt"
	localservices "http-mqtt-boilerplate/backend/internal/local/services"
	apicommon "http-mqtt-boilerplate/backend/internal/shared/api"
	"http-mqtt-boilerplate/backend/internal/shared/helpers"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
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
		queries *localdb.Queries = nil
	)

	if !config.Generate {
		// For runtime, initialize database
		err := helpers.RunMigrations(logger, config.Database, "shared/migrations", "local/migrations")
		fatalIfErr(logger, err)

		pool, err = helpers.NewPgxPool(sigCtx, logger, config.Database)
		fatalIfErr(logger, err)

		defer pool.Close()

		queries = localdb.New(pool)
	}

	// Builders
	rb, err := router.NewRouteBuilder(logger, collector)
	fatalIfErr(logger, err)

	mb, err := mqtt.NewMQTTBuilder(logger, collector, mqtt.MQTTClientOptions{
		BrokerURL: config.MQTTBroker,
		ClientID:  config.MQTTClientID,
		Username:  config.MQTTUsername,
		Password:  config.MQTTPassword,
	})
	fatalIfErr(logger, err)

	// Create services
	services := localservices.NewServices(logger, pool, queries, mb.Client())
	apiHandler := localapi.NewHandler(logger, services)
	mqttHandler := mqttapi.NewMQTTHandler(logger, services)

	registerHTTPHandlers(logger, rb, apiHandler)
	registerMQTTHandlers(logger, mb, mqttHandler)

	if config.Generate {
		// If generating, generate and exit
		if err := collector.Generate(); err != nil {
			fatalIfErr(logger, fmt.Errorf("failed to generate API documentation: %w", err))
		}

		return
	}

	go func() {
		if err := mb.Connect(sigCtx); err != nil {
			logger.Error("failed to connect to mqtt broker", utils.ErrAttr(err))
		}
	}()

	// HTTP Server
	httpAddr := fmt.Sprintf(":%d", config.Port)
	httpServer := apicommon.NewHTTPServer(logger, httpAddr, rb.Router())
	httpServer.StartOnBackground(sigCancel)

	// Wait for signal (either OS or some failure)
	<-sigCtx.Done()
	logger.Info("received signal, shutting down...")

	// Shutdown HTTP server
	logger.Info("http server shutting down...")

	if err := httpServer.ShutdownWithDefaultTimeout(); err != nil {
		logger.Error("http server shutdown failed", utils.ErrAttr(err))
	}

	logger.Info("disconnecting from MQTT broker...")
	mb.DisconnectWithDefaultTimeout()

	logger.Info("server exited gracefully")
}

// registerHTTPHandlers registers all HTTP handlers.
func registerHTTPHandlers(l *slog.Logger, rb *router.RouteBuilder, h *localapi.Handler) {
	l.Info("registering http handlers...")

	// Create middleware handler
	mw := apicommon.NewMiddlewareHandler(l)

	rb.Route("/api", func(rb *router.RouteBuilder) {
		// Add recoverer
		rb.Use(mw.RecoveryMiddleware)
		// Add request ID
		rb.Use(mw.RequestIDMiddleware)
		// Add request logger
		rb.Use(mw.LoggerMiddleware)

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

	l.Info("http handlers registered successfully")
}

// registerMQTTHandlers registers all MQTT handlers.
func registerMQTTHandlers(l *slog.Logger, mb *mqtt.MQTTBuilder, h *mqttapi.Handler) {
	l.Info("registering mqtt handlers...")
	// Telemetry operations
	h.RegisterTemperaturePublish(mb)
	h.RegisterTemperatureSubscribe(mb)
	h.RegisterSensorTelemetryPublish(mb)
	h.RegisterSensorTelemetrySubscribe(mb)

	// Control operations
	h.RegisterDeviceCommandPublish(mb)
	h.RegisterDeviceCommandSubscribe(mb)
	h.RegisterDeviceStatusPublish(mb)
	h.RegisterDeviceStatusSubscribe(mb)
	l.Info("mqtt handlers registered successfully")
}

//nolint:ireturn // Returns MetadataCollector interface (OpenAPICollector or NoopCollector)
func getCollector(c *config.Config, l *slog.Logger) (generate.MetadataCollector, error) {
	if !c.Generate {
		return &generate.NoopCollector{}, nil
	}

	return generate.NewOpenAPICollector(l, generate.OpenAPICollectorOptions{
		GoTypesDirPaths: []string{
			"backend/internal/shared/types",     // Shared types (ErrorResponse, PingResponse, etc.)
			"backend/internal/local/api/types",  // Local API types
			"backend/internal/local/mqtt/types", // Local MQTT/IoT types
		},
		DatabaseSchemaFileOutputPath: "docs/local/schema.sql",
		DocsFileOutputPath:           "docs/local/api_docs.json",
		OpenAPISpecOutputPath:        "docs/local/openapi.yaml",
		Deployment:                   "local",
		APIInfo: generate.APIInfo{
			Title:       "Local API",
			Version:     utils.GetVersionShort(),
			Description: "Local API Documentation",
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
