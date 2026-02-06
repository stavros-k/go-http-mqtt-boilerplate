package mqtt

import (
	"log/slog"

	localservices "http-mqtt-boilerplate/backend/internal/services/local"
)

// Handler handles MQTT message processing.
type Handler struct {
	l   *slog.Logger
	svc *localservices.Services
}

// NewMQTTHandler creates a new MQTT handler.
func NewMQTTHandler(l *slog.Logger, svc *localservices.Services) *Handler {
	return &Handler{
		l:   l.With(slog.String("component", "mqtt-handler")),
		svc: svc,
	}
}
