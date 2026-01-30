package mqtt

import (
	"http-mqtt-boilerplate/backend/internal/services"
	"log/slog"
)

// Handler handles MQTT message processing.
type Handler struct {
	l   *slog.Logger
	svc *services.Services
}

// NewMQTTHandler creates a new MQTT handler.
func NewMQTTHandler(l *slog.Logger, svc *services.Services) *Handler {
	return &Handler{
		l:   l.With(slog.String("component", "mqtt-handler")),
		svc: svc,
	}
}
