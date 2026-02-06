package cloud

import (
	"log/slog"

	clouddb "http-mqtt-boilerplate/backend/internal/database/clouddb/gen"

	"github.com/jackc/pgx/v5"
)

// Services holds all cloud service instances.
type Services struct {
	l       *slog.Logger
	Core    *CoreService
	queries *clouddb.Queries
}

// NewServices creates a new cloud services instance.
func NewServices(l *slog.Logger, db *pgx.Conn, queries *clouddb.Queries) *Services {
	return &Services{
		l:       l.With(slog.String("module", "cloud-services")),
		Core:    NewCoreService(l, db, queries),
		queries: queries,
	}
}
