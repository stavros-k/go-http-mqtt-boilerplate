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

	mqttbroker "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"

	"http-mqtt-boilerplate/backend/internal/config"
	localapi "http-mqtt-boilerplate/backend/internal/local/api"
	localdb "http-mqtt-boilerplate/backend/internal/local/gen"
	mqttapi "http-mqtt-boilerplate/backend/internal/local/mqtt"
	localservices "http-mqtt-boilerplate/backend/internal/local/services"
	"http-mqtt-boilerplate/backend/internal/migrations"
	sharedapi "http-mqtt-boilerplate/backend/internal/shared/api"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/migrator"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	shutdownTimeout   = 30 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	config, err := config.New()
	if err != nil {
		fatalIfErr(slog.Default(), fmt.Errorf("failed to create config: %w", err))
	}

	defer func() {
		if err := config.Close(); err != nil {
			slog.Default().Error("failed to close config", utils.ErrAttr(err))
		}
	}()

	// Initialize logger
	logger := getLogger(config)

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
		if err := runMigrations(logger, config); err != nil {
			fatalIfErr(logger, fmt.Errorf("failed to run migrations: %w", err))
		}

		pool, err = pgxpool.New(context.TODO(), config.Database)
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
		if err := mb.Connect(); err != nil {
			logger.Error("Failed to connect to MQTT broker", utils.ErrAttr(err))
		}
	}()

	//  MQTT Broker
	mqttAddr := fmt.Sprintf(":%d", config.MQTTBrokerPort)
	mqttBroker, err := getMQTTServer(logger, mqttAddr)
	fatalIfErr(logger, err)

	go func() {
		logger.Info("MQTT broker listening", slog.String("address", mqttAddr))

		if err := mqttBroker.Serve(); err != nil {
			logger.Error("MQTT broker failed", utils.ErrAttr(err))
			sigCancel()
		}
	}()

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

	logger.Info("disconnecting from MQTT broker...")
	mb.Disconnect()

	// Shutdown MQTT broker
	logger.Info("mqtt broker shutting down...")

	if err := mqttBroker.Close(); err != nil {
		logger.Error("mqtt broker shutdown failed", utils.ErrAttr(err))
	}

	logger.Info("server exited gracefully")
}

func getMQTTServer(l *slog.Logger, addr string) (*mqttbroker.Server, error) {
	server := mqttbroker.New(&mqttbroker.Options{
		Logger: l.With(slog.String("component", "mqtt-broker")),
	})
	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: addr})

	err := server.AddListener(tcp)
	if err != nil {
		return nil, err
	}

	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		return nil, err
	}

	return server, nil
}

// registerHTTPHandlers registers all HTTP handlers.
func registerHTTPHandlers(l *slog.Logger, rb *router.RouteBuilder, h *localapi.Handler) {
	l.Info("Registering HTTP handlers...")

	// Create middleware handler
	mw := sharedapi.NewMiddlewareHandler(l)

	rb.Route("/api", func(rb *router.RouteBuilder) {
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

	l.Info("HTTP handlers registered successfully")
}

// registerMQTTHandlers registers all MQTT handlers.
func registerMQTTHandlers(l *slog.Logger, mb *mqtt.MQTTBuilder, h *mqttapi.Handler) {
	l.Info("Registering MQTT handlers...")
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
	l.Info("MQTT handlers registered successfully")
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
	l.Info("Running database migrations")

	// Create migrator with shared + local migration directories
	mig, err := migrator.New(l, c.Database, migrations.GetFS(), "shared/migrations", "local/migrations")
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
