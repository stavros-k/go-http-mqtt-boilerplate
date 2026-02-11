package mqtt

import (
	"context"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const (
	keepAlive         = 30
	sendTimeout       = 10 * time.Second
	connectRetryDelay = 5 * time.Second
	connectTimeout    = 5 * time.Second
)

type MQTTClient struct {
	connMgr *autopaho.ConnectionManager
	builder *MQTTBuilder
	l       *slog.Logger
}

func newWrappedMQTTClient(l *slog.Logger, connMgr *autopaho.ConnectionManager, builder *MQTTBuilder) *MQTTClient {
	return &MQTTClient{
		connMgr: connMgr,
		builder: builder,
		l:       l.With(slog.String("component", "mqtt-client-wrapper")),
	}
}

// IsConnected returns true if the MQTT client is currently connected to the broker.
func (c *MQTTClient) IsConnected() bool {
	return c.builder.connected.Load()
}

// Publish sends a message to the specified topic using the publication spec identified by operationID.
// It does not validate the topic or payload.
// TODO: Make this easier to work with, passing operationID is a bit awkward and error prone.
func (c *MQTTClient) Publish(ctx context.Context, operationID string, actualTopic string, payload any) error {
	if c.connMgr == nil {
		return errors.New("MQTT client not connected - call Connect first")
	}

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

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	_, err = c.connMgr.Publish(ctx, &paho.Publish{
		Topic:   actualTopic,
		QoS:     byte(pub.QoS),
		Retain:  pub.Retained,
		Payload: bytes,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Warn("publish still pending")

			return errors.New("publish timeout, might complete later if reconnecting")
		}

		log.Error("publish failed", utils.ErrAttr(err))

		return err
	}

	return nil
}

// SubscribeAll subscribes to all registered subscriptions in a single call.
func (c *MQTTClient) SubscribeAll(ctx context.Context) error {
	if c.connMgr == nil {
		return errors.New("MQTT client not connected - call Connect first")
	}

	if len(c.builder.subscriptions) == 0 {
		c.l.Info("no subscriptions to subscribe to")

		return nil
	}

	// Build subscription options for all registered subscriptions
	subscriptions := make([]paho.SubscribeOptions, 0, len(c.builder.subscriptions))
	for _, sub := range c.builder.subscriptions {
		subscriptions = append(subscriptions, paho.SubscribeOptions{
			Topic: sub.TopicMQTT,
			QoS:   byte(sub.QoS),
		})
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	_, err := c.connMgr.Subscribe(ctx, &paho.Subscribe{Subscriptions: subscriptions})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("subscribe all timeout after %s, might complete later if reconnecting", sendTimeout)
		}

		return err
	}

	for _, sub := range subscriptions {
		c.l.Info("subscribed to topic", slog.String("topic", sub.Topic), slog.Int("qos", int(sub.QoS)))
	}

	c.l.Info("subscribed to all topics successfully", slog.Int("count", len(subscriptions)))

	return nil
}

// MQTTClientOptions contains configuration for creating an MQTT client.
type MQTTClientOptions struct {
	BrokerURL string
	ClientID  string
	Username  string
	Password  string
}

// newAutopahoConnection creates a new autopaho connection manager using the provided options.
func newAutopahoConnection(ctx context.Context, l *slog.Logger, opts *MQTTClientOptions, mb *MQTTBuilder) (*autopaho.ConnectionManager, error) {
	logger := l.With(
		slog.String("component", "mqtt-client"),
		slog.String("broker", opts.BrokerURL),
		slog.String("clientID", opts.ClientID),
	)
	logger.Info("creating new mqtt client with autopaho")

	// Parse broker URL
	brokerURL, err := url.Parse(opts.BrokerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse broker URL: %w", err)
	}

	// Create client config
	clientConfig := autopaho.ClientConfig{
		ServerUrls:        []*url.URL{brokerURL},
		KeepAlive:         keepAlive,
		ConnectRetryDelay: connectRetryDelay,
		ConnectTimeout:    connectTimeout,
		OnConnectionUp:    mb.onConnect(ctx),
		OnConnectionDown:  mb.onConnectionDown,
		OnConnectError:    mb.onConnectionError,
		ClientConfig: paho.ClientConfig{
			ClientID:      opts.ClientID,
			Router:        mb.router,
			OnClientError: func(err error) { logger.Error("client error", utils.ErrAttr(err)) },
		},
	}

	// Set authentication if provided
	if opts.Username != "" || opts.Password != "" {
		clientConfig.ConnectUsername = opts.Username
		clientConfig.ConnectPassword = []byte(opts.Password)
	}

	// FIXME: Set will message
	// clientConfig.WillMessage = &paho.WillMessage{
	// 	Topic:   "",
	// 	Payload: []byte(""),
	// 	QoS:     2,
	// 	Retain:  true,
	// }

	// Create connection manager
	cm, err := autopaho.NewConnection(ctx, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection manager: %w", err)
	}

	return cm, nil
}
