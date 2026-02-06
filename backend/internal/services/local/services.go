package local

import (
	"database/sql"
	"log/slog"

	sqlitegen "http-mqtt-boilerplate/backend/internal/database/sqlite/gen"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
)

// Services holds all local service instances.
type Services struct {
	l          *slog.Logger
	mqttClient *mqtt.MQTTClient
	Core       *CoreService
	queries    *sqlitegen.Queries
}

// NewServices creates a new local services instance.
func NewServices(l *slog.Logger, db *sql.DB, queries *sqlitegen.Queries, mqttClient *mqtt.MQTTClient) *Services {
	return &Services{
		l:          l.With(slog.String("module", "local-services")),
		mqttClient: mqttClient,
		Core:       NewCoreService(l, mqttClient, db, queries),
		queries:    queries,
	}
}
