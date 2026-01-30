package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/internal/api"
	"http-mqtt-boilerplate/backend/internal/config"
	sqlitegen "http-mqtt-boilerplate/backend/internal/database/sqlite/gen"
	mqttapi "http-mqtt-boilerplate/backend/internal/mqtt"
	"http-mqtt-boilerplate/backend/internal/services"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"http-mqtt-boilerplate/web"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	shutdownTimeout   = 30 * time.Second
	readHeaderTimeout = 5 * time.Second
)

func main() {
	config, err := config.New()
	if err != nil {
		fatalIfErr(slog.Default(), fmt.Errorf("failed to create config: %w", err))
	}

	defer utils.LogOnError(slog.Default(), config.Close, "failed to close config")

	// Initialize logger
	logger := getLogger(config)

	db, err := sql.Open("sqlite3", config.Database)
	fatalIfErr(logger, err)

	defer utils.LogOnError(logger, db.Close, "failed to close database")

	queries := sqlitegen.New(db)
	services := services.NewServices(logger, db, queries)
	apiHandler := api.NewAPIHandler(logger, services)
	mqttHandler := mqttapi.NewMQTTHandler(logger, services)

	// Create collector for OpenAPI generation
	collector, err := getCollector(config, logger)
	fatalIfErr(logger, err)

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

	services.RegisterMQTTClient(mb.Client())

	registerHTTPHandlers(logger, rb, apiHandler)
	registerMQTTHandlers(logger, mb, mqttHandler)

	if config.Generate {
		if err := collector.Generate(); err != nil {
			fatalIfErr(logger, fmt.Errorf("failed to generate API documentation: %w", err))
		}

		return
	}

	go func() {
		if err := mb.Connect(); err != nil {
			logger.Error("Failed to connect to MQTT broker", utils.ErrAttr(err))
		}
		defer mb.Disconnect()
	}()

	addr := fmt.Sprintf(":%d", config.Port)
	httpServer := &http.Server{
		Addr:              addr,
		Handler:           rb.Router(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	sigCtx, sigCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer sigCancel()

	go func() {
		logger.Info("http server listening", slog.String("address", addr))

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
		rb.Route("/team", func(rb *router.RouteBuilder) {
			h.RegisterGetTeam("/{teamID}", rb)
			h.RegisterPutTeam("/", rb)
			h.RegisterCreateTeam("/", rb)
			h.RegisterDeleteTeam("/", rb)
		})
	})

	web.DocsApp().Register(rb.Router(), l)
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
		GoTypesDirPath:               "backend/pkg/apitypes",
		DatabaseSchemaFileOutputPath: "schema.sql",
		DocsFileOutputPath:           "api_docs.json",
		OpenAPISpecOutputPath:        "openapi.yaml",
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
