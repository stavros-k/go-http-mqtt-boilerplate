package api

import (
	"log/slog"

	cloudservices "http-mqtt-boilerplate/backend/internal/cloud/services"
)

const (
	CoreGroup = "Core"
	TeamGroup = "Team"
)

// Handler represents the cloud API handler.
type Handler struct {
	l   *slog.Logger
	svc *cloudservices.Services
}

// NewHandler creates a new cloud API handler.
func NewHandler(l *slog.Logger, svc *cloudservices.Services) *Handler {
	return &Handler{
		l:   l.With(slog.String("component", "cloudapi")),
		svc: svc,
	}
}
