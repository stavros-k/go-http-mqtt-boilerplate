package api

import (
	"log/slog"

	localservices "http-mqtt-boilerplate/backend/internal/local/services"
)

const (
	CoreGroup = "Core"
	TeamGroup = "Team"
)

// Handler represents the local API handler.
type Handler struct {
	l   *slog.Logger
	svc *localservices.Services
}

// NewHandler creates a new local API handler.
func NewHandler(l *slog.Logger, svc *localservices.Services) *Handler {
	return &Handler{
		l:   l.With(slog.String("component", "localapi")),
		svc: svc,
	}
}
