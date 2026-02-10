package mqtt

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/eclipse/paho.golang/paho"

	"http-mqtt-boilerplate/backend/internal/local/mqtt/types"
	"http-mqtt-boilerplate/backend/pkg/mqtt"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

// RegisterDeviceCommandPublish registers the device command publication operation.
func (s *Handler) RegisterDeviceCommandPublish(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterPublish("devices/{deviceID}/commands", mqtt.PublicationSpec{
		OperationID: "publishDeviceCommand",
		Summary:     "Publish device command",
		Description: "Sends commands to IoT devices.",
		Group:       "Control",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Unique identifier of the target device",
				Type:        new(string),
			},
		},
		MessageType: types.DeviceCommand{
			DeviceID: "device-001",
			Command:  "restart",
		},
		QoS:      mqtt.QoSAtLeastOnce,
		Retained: false,
		Examples: map[string]any{
			"restart": types.DeviceCommand{
				DeviceID: "device-001",
				Command:  "restart",
			},
			"updateConfig": types.DeviceCommand{
				DeviceID: "device-001",
				Command:  "update_config",
				Parameters: map[string]string{
					"interval": "60",
					"enabled":  "true",
				},
			},
		},
	})
}

// RegisterDeviceCommandSubscribe registers the device command subscription operation.
func (s *Handler) RegisterDeviceCommandSubscribe(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterSubscribe("devices/{deviceID}/commands", mqtt.SubscriptionSpec{
		OperationID: "subscribeDeviceCommand",
		Summary:     "Subscribe to device commands",
		Description: "Receives commands sent to IoT devices for logging and monitoring.",
		Group:       "Control",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Matches any device ID",
				Type:        new(string),
			},
		},
		MessageType: types.DeviceCommand{
			DeviceID: "device-001",
			Command:  "restart",
		},
		Handler: s.handleDeviceCommand,
		QoS:     mqtt.QoSAtLeastOnce,
		Examples: map[string]any{
			"restart": types.DeviceCommand{
				DeviceID: "device-001",
				Command:  "restart",
			},
		},
	})
}

// handleDeviceCommand handles incoming device commands.
func (s *Handler) handleDeviceCommand(msg *paho.Publish) {
	var command types.DeviceCommand
	if err := json.Unmarshal(msg.Payload, &command); err != nil {
		s.l.Error("failed to unmarshal device command",
			slog.String("topic", msg.Topic),
			utils.ErrAttr(err))

		return
	}

	s.l.Info("received device command",
		slog.String("deviceID", command.DeviceID),
		slog.String("command", command.Command),
		slog.Any("parameters", command.Parameters))

	// Process the command (e.g., log, validate, forward to device)
	// TODO: Add your business logic here
}

// RegisterDeviceStatusPublish registers the device status publication operation.
func (s *Handler) RegisterDeviceStatusPublish(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterPublish("devices/{deviceID}/status", mqtt.PublicationSpec{
		OperationID: "publishDeviceStatus",
		Summary:     "Publish device status",
		Description: "Publishes device status updates.",
		Group:       "Control",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Unique identifier of the device",
				Type:        new(string),
			},
		},
		MessageType: types.DeviceStatus{
			DeviceID:  "device-001",
			Status:    "online",
			Uptime:    3600,
			Timestamp: time.Time{},
		},
		QoS:      mqtt.QoSAtLeastOnce,
		Retained: true,
		Examples: map[string]any{
			"online": types.DeviceStatus{
				DeviceID:  "device-001",
				Status:    "online",
				Uptime:    3600,
				Timestamp: time.Time{},
			},
			"offline": types.DeviceStatus{
				DeviceID:  "device-001",
				Status:    "offline",
				Uptime:    0,
				Timestamp: time.Time{},
			},
		},
	})
}

// RegisterDeviceStatusSubscribe registers the device status subscription operation.
func (s *Handler) RegisterDeviceStatusSubscribe(mb *mqtt.MQTTBuilder) {
	mb.MustRegisterSubscribe("devices/{deviceID}/status", mqtt.SubscriptionSpec{
		OperationID: "subscribeDeviceStatus",
		Summary:     "Subscribe to device status",
		Description: "Receives device status updates from all IoT devices.",
		Group:       "Control",
		TopicParameters: []mqtt.TopicParameter{
			{
				Name:        "deviceID",
				Description: "Matches any device ID",
				Type:        new(string),
			},
		},
		MessageType: types.DeviceStatus{
			DeviceID:  "device-001",
			Status:    "online",
			Uptime:    3600,
			Timestamp: time.Time{},
		},
		Handler: s.handleDeviceStatus,
		QoS:     mqtt.QoSAtLeastOnce,
		Examples: map[string]any{
			"online": types.DeviceStatus{
				DeviceID:  "device-001",
				Status:    "online",
				Uptime:    3600,
				Timestamp: time.Time{},
			},
		},
	})
}

// handleDeviceStatus handles incoming device status updates.
func (s *Handler) handleDeviceStatus(msg *paho.Publish) {
	var status types.DeviceStatus
	if err := json.Unmarshal(msg.Payload, &status); err != nil {
		s.l.Error("failed to unmarshal device status",
			slog.String("topic", msg.Topic),
			utils.ErrAttr(err))

		return
	}

	s.l.Info("received device status",
		slog.String("deviceID", status.DeviceID),
		slog.String("status", status.Status),
		slog.Int64("uptime", status.Uptime))

	// Process the status (e.g., update device registry, trigger alerts)
	// TODO: Add your business logic here
}
