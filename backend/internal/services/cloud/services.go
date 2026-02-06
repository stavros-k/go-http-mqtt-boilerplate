package cloud

import (
	"log/slog"

	"github.com/jackc/pgx/v5"

	postgresgen "http-mqtt-boilerplate/backend/internal/database/postgres/gen"
)

// Services holds all cloud service instances.
type Services struct {
	l       *slog.Logger
	Core    *CoreService
	queries *postgresgen.Queries
}

// NewServices creates a new cloud services instance.
func NewServices(l *slog.Logger, db *pgx.Conn, queries *postgresgen.Queries) *Services {
	return &Services{
		l:       l.With(slog.String("module", "cloud-services")),
		Core:    NewCoreService(l, db, queries),
		queries: queries,
	}
}
