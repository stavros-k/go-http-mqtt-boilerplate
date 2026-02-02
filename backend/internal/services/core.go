package services

import (
	"context"
	"database/sql"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"log/slog"
)

type CoreService struct {
	l    *slog.Logger
	mqtt *mqtt.MQTTClient
	db   *sql.DB
}

func NewCoreService(l *slog.Logger, mqttClient *mqtt.MQTTClient, db *sql.DB) *CoreService {
	return &CoreService{
		l:    l.With(slog.String("service", "core")),
		mqtt: mqttClient,
		db:   db,
	}
}

type HealthStatus struct {
	Database bool
	MQTT     bool
}

func (s *CoreService) Health(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Database: true,
		MQTT:     true,
	}

	if err := s.db.PingContext(ctx); err != nil {
		s.l.Error("database unreachable", slog.String("error", err.Error()))
		status.Database = false
	}

	if !s.mqtt.IsConnected() {
		s.l.Error("mqtt broker unreachable")
		status.MQTT = false
	}

	return status
}
