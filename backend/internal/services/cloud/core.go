package cloud

import (
	"context"
	"log/slog"

	clouddb "http-mqtt-boilerplate/backend/internal/database/clouddb/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CoreService handles core business logic for the cloud API.
type CoreService struct {
	l  *slog.Logger
	db *pgxpool.Pool
	q  *clouddb.Queries
}

// NewCoreService creates a new core service instance.
func NewCoreService(l *slog.Logger, db *pgxpool.Pool, queries *clouddb.Queries) *CoreService {
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
