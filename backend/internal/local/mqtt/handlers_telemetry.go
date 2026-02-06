package mqtt

import (
	"encoding/json"
	"log/slog"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"

	"http-mqtt-boilerplate/backend/internal/local/mqtt/types"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

// RegisterTemperaturePublish registers the temperature publication operation.
func (s *Handler) RegisterTemperaturePublish(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterPublish("devices/{deviceID}/temperature", mqtt.PublicationSpec{
		OperationID: "publishTemperature",
		Summary:     "Publish temperature reading",
		Description: "Publishes temperature readings from IoT devices. The device ID is part of the topic path.",
		Group:       "Telemetry",
		Deprecated:  "This will be removed in the feature.",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Unique identifier of the device sending the temperature reading",
				Type:        new(string),
			},
		},
		MessageType: types.TemperatureReading{
			DeviceID:    "device-001",
			Temperature: 22.5,
			Unit:        "celsius",
			Timestamp:   time.Time{},
		},
		QoS:      mqtt.QoSAtLeastOnce,
		Retained: true,
		Examples: map[string]any{
			"normal": types.TemperatureReading{
				DeviceID:    "device-001",
				Temperature: 22.5,
				Unit:        "celsius",
				Timestamp:   time.Time{},
			},
			"fahrenheit": types.TemperatureReading{
				DeviceID:    "device-002",
				Temperature: 72.5,
				Unit:        "fahrenheit",
				Timestamp:   time.Time{},
			},
		},
	})
}

// RegisterTemperatureSubscribe registers the temperature subscription operation.
func (s *Handler) RegisterTemperatureSubscribe(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterSubscribe("devices/{deviceID}/temperature", mqtt.SubscriptionSpec{
		OperationID: "subscribeTemperature",
		Summary:     "Subscribe to temperature readings",
		Description: "Receives temperature readings from all IoT devices.",
		Group:       "Telemetry",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Matches any device ID",
				Type:        new(string),
			},
		},
		MessageType: types.TemperatureReading{
			DeviceID:    "device-001",
			Temperature: 22.5,
			Unit:        "celsius",
			Timestamp:   time.Time{},
		},
		Handler: s.handleTemperature,
		QoS:     mqtt.QoSAtLeastOnce,
		Examples: map[string]any{
			"normal": types.TemperatureReading{
				DeviceID:    "device-001",
				Temperature: 22.5,
				Unit:        "celsius",
				Timestamp:   time.Time{},
			},
		},
	})
}

// handleTemperature handles incoming temperature readings.
func (s *Handler) handleTemperature(client pahomqtt.Client, msg pahomqtt.Message) {
	var reading types.TemperatureReading
	if err := json.Unmarshal(msg.Payload(), &reading); err != nil {
		s.l.Error("Failed to unmarshal temperature reading", slog.String("topic", msg.Topic()), utils.ErrAttr(err))

		return
	}

	s.l.Info("Received temperature reading", slog.String("deviceID", reading.DeviceID), slog.Float64("temperature", reading.Temperature), slog.String("unit", reading.Unit), slog.Time("timestamp", reading.Timestamp))

	// Process the reading (e.g., store in database, trigger alerts, etc.)
	// TODO: Add your business logic here
}

// RegisterSensorTelemetryPublish registers the sensor telemetry publication operation.
func (s *Handler) RegisterSensorTelemetryPublish(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterPublish("devices/{deviceID}/sensors/{sensorType}", mqtt.PublicationSpec{
		OperationID: "publishSensorTelemetry",
		Summary:     "Publish sensor telemetry",
		Description: "Publishes generic sensor telemetry data from IoT devices.",
		Group:       "Telemetry",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Unique identifier of the device",
				Type:        new(string),
			},
			{
				Name:        "sensorType",
				Description: "Type of sensor (e.g., humidity, pressure, motion)",
				Type:        new(string),
			},
		},
		MessageType: types.SensorTelemetry{
			DeviceID:   "device-001",
			SensorType: "humidity",
			Value:      65.5,
			Unit:       "percent",
			Timestamp:  time.Time{},
			Quality:    95,
		},
		QoS:      mqtt.QoSAtLeastOnce,
		Retained: false,
		Examples: map[string]any{
			"humidity": types.SensorTelemetry{
				DeviceID:   "device-001",
				SensorType: "humidity",
				Value:      65.5,
				Unit:       "percent",
				Timestamp:  time.Time{},
				Quality:    95,
			},
			"pressure": types.SensorTelemetry{
				DeviceID:   "device-001",
				SensorType: "pressure",
				Value:      1013.25,
				Unit:       "hPa",
				Timestamp:  time.Time{},
				Quality:    100,
			},
		},
	})
}

// RegisterSensorTelemetrySubscribe registers the sensor telemetry subscription operation.
func (s *Handler) RegisterSensorTelemetrySubscribe(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterSubscribe("devices/{deviceID}/sensors/{sensorType}", mqtt.SubscriptionSpec{
		OperationID: "subscribeSensorTelemetry",
		Summary:     "Subscribe to sensor telemetry",
		Description: "Receives generic sensor telemetry data from all IoT devices and sensor types.",
		Group:       "Telemetry",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Matches any device ID",
				Type:        new(string),
			},
			{
				Name:        "sensorType",
				Description: "Matches any sensor type",
				Type:        new(string),
			},
		},
		MessageType: types.SensorTelemetry{
			DeviceID:   "device-001",
			SensorType: "humidity",
			Value:      65.5,
			Unit:       "percent",
			Timestamp:  time.Time{},
			Quality:    95,
		},
		Handler: s.handleSensorTelemetry,
		QoS:     mqtt.QoSAtLeastOnce,
		Examples: map[string]any{
			"humidity": types.SensorTelemetry{
				DeviceID:   "device-001",
				SensorType: "humidity",
				Value:      65.5,
				Unit:       "percent",
				Timestamp:  time.Time{},
				Quality:    95,
			},
		},
	})
}

// handleSensorTelemetry handles incoming sensor telemetry data.
func (s *Handler) handleSensorTelemetry(client pahomqtt.Client, msg pahomqtt.Message) {
	var telemetry types.SensorTelemetry
	if err := json.Unmarshal(msg.Payload(), &telemetry); err != nil {
		s.l.Error("Failed to unmarshal sensor telemetry", slog.String("topic", msg.Topic()), utils.ErrAttr(err))

		return
	}

	s.l.Info("Received sensor telemetry", slog.String("deviceID", telemetry.DeviceID), slog.String("sensorType", telemetry.SensorType), slog.Float64("value", telemetry.Value), slog.String("unit", telemetry.Unit), slog.Int("quality", telemetry.Quality))

	// Process the telemetry (e.g., store in database, trigger alerts, etc.)
	// TODO: Add your business logic here
}
