package mqtt

import (
	"context"
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const (
	disconnectTimeout = 10 * time.Second
)

// MQTTBuilder provides a fluent API for registering MQTT publications and subscriptions.
type MQTTBuilder struct {
	connMgr       *autopaho.ConnectionManager
	wrappedClient *MQTTClient
	collector     generate.MQTTMetadataCollector
	router        *paho.StandardRouter
	l             *slog.Logger
	operationIDs  map[string]struct{}
	publications  map[string]*PublicationSpec
	subscriptions map[string]*SubscriptionSpec
	connected     atomic.Bool
	opts          MQTTClientOptions

	registrationsCompleted atomic.Bool
}

// NewMQTTBuilder creates a new MQTT builder with the given broker configuration.
// The builder does not connect immediately; call [MQTTBuilder.Connect] after registering all subscriptions.
func NewMQTTBuilder(l *slog.Logger, collector generate.MQTTMetadataCollector, opts MQTTClientOptions) (*MQTTBuilder, error) {
	mqttBuilderLogger := l.With(slog.String("component", "mqtt-builder"))

	if opts.BrokerURL == "" {
		return nil, errors.New("broker URL is required")
	}

	if opts.ClientID == "" {
		return nil, errors.New("client ID is required")
	}

	// Create a router for handling incoming messages
	router := paho.NewStandardRouter()

	mb := &MQTTBuilder{
		collector:     collector,
		router:        router,
		l:             mqttBuilderLogger,
		opts:          opts,
		operationIDs:  make(map[string]struct{}),
		publications:  make(map[string]*PublicationSpec),
		subscriptions: make(map[string]*SubscriptionSpec),
	}

	// Create wrapped client with nil connMgr - will be populated in [MQTTBuilder.Connect]
	// This allows [MQTTBuilder.Client] to be called before [MQTTBuilder.Connect]
	mb.wrappedClient = newWrappedMQTTClient(l, nil, mb)

	mqttBuilderLogger.Info("mqtt builder created", slog.String("broker", opts.BrokerURL), slog.String("clientID", opts.ClientID))

	return mb, nil
}

// Client returns the underlying MQTT client.
func (mb *MQTTBuilder) Client() *MQTTClient {
	return mb.wrappedClient
}

// RegisterPublish registers a publication operation.
func (mb *MQTTBuilder) RegisterPublish(topic string, spec PublicationSpec) error {
	if mb.registrationsCompleted.Load() {
		return errors.New("cannot register publication after connecting to MQTT broker")
	}

	// Validate topic
	if err := validateTopicPattern(topic); err != nil {
		return fmt.Errorf("invalid topic pattern: %w", err)
	}

	// Validate spec
	if err := mb.validatePublicationSpec(spec); err != nil {
		return fmt.Errorf("invalid publication spec: %w", err)
	}

	// Check for duplicate operationID
	if _, exists := mb.operationIDs[spec.OperationID]; exists {
		return fmt.Errorf("duplicate operationID: %s", spec.OperationID)
	}

	// Convert topic parameters to documentation format
	topicParams, err := generateParameters(topic, spec.TopicParameters)
	if err != nil {
		return fmt.Errorf("failed to generate topic parameters in operationID %s: %w", spec.OperationID, err)
	}

	// Convert parameterized topic to MQTT wildcard format
	mqttTopic := convertTopicToMQTT(topic)
	spec.TopicMQTT = mqttTopic

	// Register with collector
	if err := mb.collector.RegisterMQTTPublication(&generate.MQTTPublicationInfo{
		OperationID:     spec.OperationID,
		Topic:           topic,
		TopicMQTT:       mqttTopic,
		TopicParameters: topicParams,
		Summary:         spec.Summary,
		Description:     spec.Description,
		Group:           spec.Group,
		Deprecated:      spec.Deprecated,
		QoS:             byte(spec.QoS),
		Retained:        spec.Retained,
		TypeValue:       spec.MessageType,
		Examples:        spec.Examples,
	}); err != nil {
		return fmt.Errorf("failed to register publication with collector: %w", err)
	}

	// Store publication
	mb.operationIDs[spec.OperationID] = struct{}{}
	mb.publications[spec.OperationID] = &spec

	mb.l.Info("Registered MQTT publication", slog.String("operationID", spec.OperationID), slog.String("topic", topic), slog.String("group", spec.Group))

	return nil
}

// MustRegisterPublish registers a publication operation and terminates the program if an error occurs.
func (mb *MQTTBuilder) MustRegisterPublish(topic string, spec PublicationSpec) {
	if err := mb.RegisterPublish(topic, spec); err != nil {
		mb.l.Error("failed to register publication", slog.String("operationID", spec.OperationID), slog.String("topic", topic), slog.String("group", spec.Group), utils.ErrAttr(err))
		os.Exit(1)
	}
}

// RegisterSubscribe registers a subscription operation.
func (mb *MQTTBuilder) RegisterSubscribe(topic string, spec SubscriptionSpec) error {
	if mb.registrationsCompleted.Load() {
		return errors.New("cannot register subscription after connecting to MQTT broker")
	}

	sanitizedTopic := generate.SanitizePath(topic)
	if topic != sanitizedTopic {
		return fmt.Errorf("invalid topic pattern: topic %q does not match sanitized form %q", topic, sanitizedTopic)
	}

	// Validate topic
	if err := validateTopicPattern(topic); err != nil {
		return fmt.Errorf("invalid topic pattern: %w", err)
	}

	// Validate spec
	if err := mb.validateSubscriptionSpec(spec); err != nil {
		return fmt.Errorf("invalid subscription spec: %w", err)
	}

	// Check for duplicate operationID
	if _, exists := mb.operationIDs[spec.OperationID]; exists {
		return fmt.Errorf("duplicate operationID: %s", spec.OperationID)
	}

	// Generate topic parameters
	topicParams, err := generateParameters(topic, spec.TopicParameters)
	if err != nil {
		return fmt.Errorf("failed to generate topic parameters in operationID %s: %w", spec.OperationID, err)
	}

	// Convert parameterized topic to MQTT wildcard format
	mqttTopic := convertTopicToMQTT(topic)
	spec.TopicMQTT = mqttTopic

	// Register with collector
	if err := mb.collector.RegisterMQTTSubscription(&generate.MQTTSubscriptionInfo{
		OperationID:     spec.OperationID,
		Topic:           topic,
		TopicMQTT:       mqttTopic,
		TopicParameters: topicParams,
		Summary:         spec.Summary,
		Description:     spec.Description,
		Group:           spec.Group,
		Deprecated:      spec.Deprecated,
		QoS:             byte(spec.QoS),
		TypeValue:       spec.MessageType,
		Examples:        spec.Examples,
	}); err != nil {
		return fmt.Errorf("failed to register subscription with collector: %w", err)
	}

	// Store subscription with MQTT wildcard topic (for actual subscription)
	mb.operationIDs[spec.OperationID] = struct{}{}
	mb.subscriptions[spec.OperationID] = &spec

	// Register handler with the router
	mb.router.RegisterHandler(mqttTopic, spec.Handler)

	mb.l.Info("Registered MQTT subscription", slog.String("operationID", spec.OperationID), slog.String("topic", topic), slog.String("group", spec.Group))

	return nil
}

// MustRegisterSubscribe registers a subscription operation and terminates the program if an error occurs.
func (mb *MQTTBuilder) MustRegisterSubscribe(topic string, spec SubscriptionSpec) {
	if err := mb.RegisterSubscribe(topic, spec); err != nil {
		mb.l.Error("failed to register subscription", slog.String("operationID", spec.OperationID), slog.String("topic", topic), slog.String("group", spec.Group), utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Connect connects to the MQTT broker and waits for the connection to complete.
// This will disallow any further registration calls.
// [MQTTBuilder.RegisterPublish], [MQTTBuilder.MustRegisterPublish],[MQTTBuilder.RegisterSubscribe], [MQTTBuilder.MustRegisterSubscribe].
func (mb *MQTTBuilder) Connect(ctx context.Context) error {
	mb.registrationsCompleted.Store(true)

	// Create the autopaho connection now that all registrations are complete
	connMgr, err := newAutopahoConnection(ctx, mb.l, &mb.opts, mb)
	if err != nil {
		return fmt.Errorf("failed to create connection manager: %w", err)
	}

	mb.connMgr = connMgr
	mb.wrappedClient.connMgr = connMgr

	mb.l.Info("Connecting to MQTT broker... Will wait indefinitely for connection to complete")

	done := make(chan struct{})
	defer close(done)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if mb.wrappedClient.IsConnected() {
					return
				}

				mb.l.Warn("mqtt has not completed an initial connection yet, still waiting...")
			}
		}
	}()

	// Wait for the initial connection with autopaho
	// autopaho will automatically connect when [autopaho.NewConnection] is called
	// from within the [newAutopahoConnection] function
	// We just need to wait for the first successful connection
	err = mb.connMgr.AwaitConnection(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	mb.l.Info("Connection to MQTT broker established")

	return nil
}

// DisconnectWithDefaultTimeout disconnects from the MQTT broker with a default timeout.
func (mb *MQTTBuilder) DisconnectWithDefaultTimeout() {
	if !mb.wrappedClient.IsConnected() {
		return
	}

	mb.l.Info("Disconnecting from MQTT broker...")

	ctx, cancel := context.WithTimeout(context.Background(), disconnectTimeout)
	defer cancel()

	// Send disconnect packet
	err := mb.connMgr.Disconnect(ctx)
	if err != nil {
		mb.l.Error("failed to disconnect from mqtt broker", utils.ErrAttr(err))
		return
	}

	mb.l.Info("Disconnected from MQTT broker")
}

// onConnect is called when the client successfully connects or reconnects to the broker.
func (mb *MQTTBuilder) onConnect(ctx context.Context) func(*autopaho.ConnectionManager, *paho.Connack) {
	return func(_ *autopaho.ConnectionManager, _ *paho.Connack) {
		mb.l.Info("Connected to MQTT broker, subscribing to topics", slog.Int("subscriptionCount", len(mb.subscriptions)))
		mb.connected.Store(true)
		// Subscribe to all registered subscriptions at once
		go func() {
			if err := mb.wrappedClient.SubscribeAll(ctx); err != nil {
				mb.l.Error("failed to subscribe to topics", utils.ErrAttr(err))
			}
		}()
	}
}

// onConnectionError is called when the client fails to connect to the broker.
func (mb *MQTTBuilder) onConnectionError(err error) {
	mb.l.Warn("failed to connect to mqtt broker", utils.ErrAttr(err))
}

// onConnectionDown is called when an active connection to the broker is lost.
func (mb *MQTTBuilder) onConnectionDown() bool {
	mb.l.Warn("connection to mqtt broker lost")
	mb.connected.Store(false)
	return true // Return true to allow autopaho to attempt reconnection
}
