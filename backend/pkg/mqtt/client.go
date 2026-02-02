package mqtt

import (
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client  mqtt.Client
	builder *MQTTBuilder
}

// Publish sends a message to the specified topic using the publication spec identified by operationID.
// It does not validate the topic or payload.
func (c *MQTTClient) Publish(operationID string, actualTopic string, payload any) error {
	pub, ok := c.builder.publications[operationID]
	if !ok {
		return fmt.Errorf("publication not found for operationID %s", operationID)
	}

	bytes, err := utils.ToJSON(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	token := c.client.Publish(actualTopic, byte(pub.QoS), pub.Retained, bytes)
	token.Wait()

	if err := token.Error(); err != nil {
		return fmt.Errorf("failed to publish to topic %s: %w", actualTopic, err)
	}

	return nil
}

func (c *MQTTClient) IsConnected() bool {
	return c.client.IsConnectionOpen()
}

// MQTTClientOptions contains configuration for creating an MQTT client.
type MQTTClientOptions struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
}

func newMQTTClient(l *slog.Logger, opts *MQTTClientOptions, mb *MQTTBuilder) mqtt.Client {
	logger := l.With(
		slog.String("component", "mqtt-client"),
		slog.String("broker", opts.BrokerURL),
		slog.String("clientID", opts.ClientID),
	)
	logger.Info("Creating new MQTT client")
	// TODO: Check this
	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(opts.BrokerURL)
	clientOpts.SetClientID(opts.ClientID)

	if opts.Username != "" {
		clientOpts.SetUsername(opts.Username)
	}

	if opts.Password != "" {
		clientOpts.SetPassword(opts.Password)
	}

	// Retry every 5 seconds, max interval 15 seconds
	clientOpts.SetAutoReconnect(true)
	clientOpts.SetConnectRetry(true)
	clientOpts.SetConnectTimeout(5 * time.Second)
	clientOpts.SetConnectRetryInterval(5 * time.Second)
	clientOpts.SetMaxReconnectInterval(15 * time.Second)
	clientOpts.SetKeepAlive(30 * time.Second)

	// Set connection callbacks
	clientOpts.SetOnConnectHandler(mb.onConnect)
	clientOpts.SetConnectionLostHandler(mb.onConnectionLost)
	clientOpts.SetReconnectingHandler(mb.onReconnecting)
	// FIXME: Uncomment this on next release
	// clientOpts.SetLogger(logger)
	// FIXME: Set will message
	// clientOpts.SetWill("", "", 2, true)

	return mqtt.NewClient(clientOpts)
}
