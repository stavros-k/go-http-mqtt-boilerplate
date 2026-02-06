package local

import (
	"database/sql"
	"log/slog"

	localdb "http-mqtt-boilerplate/backend/internal/database/localdb/gen"
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
func NewServices(l *slog.Logger, db *sql.DB, queries *localdb.Queries, mqttClient *mqtt.MQTTClient) *Services {
	return &Services{
		l:          l.With(slog.String("module", "local-services")),
		mqttClient: mqttClient,
		Core:       NewCoreService(l, mqttClient, db, queries),
		queries:    queries,
	}
}
