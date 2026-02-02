package mqtt

import (
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client  mqtt.Client
	builder *MQTTBuilder
	l       *slog.Logger
}

func newWrappedMQTTClient(l *slog.Logger, client mqtt.Client, builder *MQTTBuilder) *MQTTClient {
	return &MQTTClient{
		client:  client,
		builder: builder,
		l:       l.With(slog.String("component", "mqtt-client-wrapper")),
	}
}

// IsConnected returns true if the MQTT client is currently connected to the broker.
func (c *MQTTClient) IsConnected() bool {
	return c.client.IsConnectionOpen()
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

	log := c.l.With(
		slog.String("operationID", pub.OperationID),
		slog.String("topic", actualTopic),
		slog.Int("qos", int(pub.QoS)),
	)
	token := c.client.Publish(actualTopic, byte(pub.QoS), pub.Retained, bytes)

	go func() {
		if !token.WaitTimeout(30 * time.Second) {
			log.Warn("Publish still pending after 30s", slog.Int("qos", int(pub.QoS)))

			return
		}

		if err := token.Error(); err != nil {
			log.Error("Publish failed", utils.ErrAttr(err))
		}
	}()

	return nil
}

func (c *MQTTClient) Subscribe(operationID string) error {
	sub, ok := c.builder.subscriptions[operationID]
	if !ok {
		return fmt.Errorf("subscription not found for operationID %s", operationID)
	}

	log := c.l.With(
		slog.String("operationID", sub.OperationID),
		slog.String("topic", sub.TopicMQTT),
		slog.Int("qos", int(sub.QoS)),
	)

	token := c.client.Subscribe(sub.TopicMQTT, byte(sub.QoS), sub.Handler)
	if !token.WaitTimeout(10 * time.Second) {
		return errors.New("subscribe timeout")
	}

	if err := token.Error(); err != nil {
		return err
	}

	log.Info("Subscribed successfully")

	return nil
}

// MQTTClientOptions contains configuration for creating an MQTT client.
type MQTTClientOptions struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
}

// newLowLevelMQTTClient creates a new low-level MQTT client using the provided options.
//
//nolint:ireturn // We return the library MQTT client type
func newLowLevelMQTTClient(l *slog.Logger, opts *MQTTClientOptions, mb *MQTTBuilder) mqtt.Client {
	logger := l.With(
		slog.String("component", "mqtt-client"),
		slog.String("broker", opts.BrokerURL),
		slog.String("clientID", opts.ClientID),
	)
	logger.Info("Creating new MQTT client")

	clientOpts := mqtt.NewClientOptions()
	// FIXME: Uncomment this on next release
	// clientOpts.SetLogger(logger)

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
	// FIXME: Set will message
	// clientOpts.SetWill("", "", 2, true)

	return mqtt.NewClient(clientOpts)
}
