package local

import (
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	localdb "http-mqtt-boilerplate/backend/internal/local/gen"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
)

// Services holds all local service instances.
type Services struct {
	l          *slog.Logger
	mqttClient *mqtt.MQTTClient
	Core       *CoreService
	queries    *localdb.Queries
}

// NewServices creates a new local services instance.
func NewServices(l *slog.Logger, pool *pgxpool.Pool, queries *localdb.Queries, mqttClient *mqtt.MQTTClient) *Services {
	return &Services{
		l:          l.With(slog.String("module", "local-services")),
		mqttClient: mqttClient,
		Core:       NewCoreService(l, mqttClient, pool, queries),
		queries:    queries,
	}
}
