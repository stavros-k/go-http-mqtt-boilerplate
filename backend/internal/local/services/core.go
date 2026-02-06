package local

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	localdb "http-mqtt-boilerplate/backend/internal/local/gen"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
)

// CoreService handles core business logic for the local API.
type CoreService struct {
	l    *slog.Logger
	mqtt *mqtt.MQTTClient
	pool *pgxpool.Pool
	q    *localdb.Queries
}

// NewCoreService creates a new core service instance.
func NewCoreService(l *slog.Logger, mqttClient *mqtt.MQTTClient, pool *pgxpool.Pool, queries *localdb.Queries) *CoreService {
	return &CoreService{
		l:    l.With(slog.String("service", "core")),
		mqtt: mqttClient,
		pool: pool,
		q:    queries,
	}
}

// HealthStatus represents the health status of local services.
type HealthStatus struct {
	Database bool
	MQTT     bool
}

// Health checks the health of local services (database and MQTT).
func (s *CoreService) Health(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Database: true,
		MQTT:     true,
	}

	if err := s.pool.Ping(ctx); err != nil {
		s.l.Error("database unreachable", slog.String("error", err.Error()))

		status.Database = false
	}

	if !s.mqtt.IsConnected() {
		s.l.Error("mqtt broker unreachable")

		status.MQTT = false
	}

	return status
}
