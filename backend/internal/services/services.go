package services

import (
	"database/sql"
	sqlitegen "http-mqtt-boilerplate/backend/internal/database/sqlite/gen"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"log/slog"
)

type Services struct {
	l          *slog.Logger
	mqttClient *mqtt.MQTTClient
	Core       *CoreService
}

func NewServices(l *slog.Logger, db *sql.DB, queries *sqlitegen.Queries) *Services {
	return &Services{
		l:    l.With(slog.String("module", "services")),
		Core: NewCoreService(l, db),
	}
}

// RegisterMQTTClient registers the MQTT client with the services.
func (s *Services) RegisterMQTTClient(client *mqtt.MQTTClient) {
	s.mqttClient = client
}
