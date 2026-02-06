package cloud

import (
	"context"
	postgresgen "http-mqtt-boilerplate/backend/internal/database/postgres/gen"
	"log/slog"

	"github.com/jackc/pgx/v5"
)

// CoreService handles core business logic for the cloud API.
type CoreService struct {
	l  *slog.Logger
	db *pgx.Conn
	q  *postgresgen.Queries
}

// NewCoreService creates a new core service instance.
func NewCoreService(l *slog.Logger, db *pgx.Conn, queries *postgresgen.Queries) *CoreService {
	return &CoreService{
		l:  l.With(slog.String("service", "core")),
		db: db,
		q:  queries,
	}
}

// HealthStatus represents the health status of cloud services.
type HealthStatus struct {
	Database bool
}

// Health checks the health of cloud services (database only, no MQTT).
func (s *CoreService) Health(ctx context.Context) HealthStatus {
	status := HealthStatus{
		Database: true,
	}

	if err := s.db.Ping(ctx); err != nil {
		s.l.Error("database unreachable", slog.String("error", err.Error()))

		status.Database = false
	}

	return status
}
