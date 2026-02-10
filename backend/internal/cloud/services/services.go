package cloud

import (
	"log/slog"

	clouddb "http-mqtt-boilerplate/backend/internal/cloud/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Services holds all cloud service instances.
type Services struct {
	l       *slog.Logger
	Core    *CoreService
	queries *clouddb.Queries
}

// NewServices creates a new cloud services instance.
func NewServices(l *slog.Logger, db *pgxpool.Pool, queries *clouddb.Queries) *Services {
	return &Services{
		l:       l.With(slog.String("module", "cloud-services")),
		Core:    NewCoreService(l, db, queries),
		queries: queries,
	}
}
