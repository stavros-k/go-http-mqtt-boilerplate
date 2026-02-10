package mqtt

import (
	"github.com/eclipse/paho.golang/paho"
)

// QoS represents MQTT quality of service levels.
type QoS byte

const (
	// QoSAtMostOnce means the message is delivered at most once, or it may not be delivered at all.
	QoSAtMostOnce QoS = 0
	// QoSAtLeastOnce means the message is always delivered at least once.
	QoSAtLeastOnce QoS = 1
	// QoSExactlyOnce means the message is always delivered exactly once.
	QoSExactlyOnce QoS = 2
)

// TopicParameter describes a parameter in an MQTT topic pattern.
type TopicParameter struct {
	Name        string // Name is the parameter name (e.g., "deviceID")
	Description string // Description explains what this parameter represents
	Type        any    // Type is the Go type of the parameter (e.g., new(string))
}

// PublicationSpec describes an MQTT publication operation.
type PublicationSpec struct {
	OperationID     string           // OperationID is a unique identifier for this publication operation (e.g., "publishTemperature").
	TopicMQTT       string           // TopicMQTT is the MQTT wildcard format (e.g., devices/+/temperature).
	Summary         string           // Summary is a short description of the publication.
	Description     string           // Description provides detailed information about the publication.
	Group           string           // Group is a logical grouping for the publication (e.g., "Telemetry", "Control").
	Deprecated      string           // Deprecated contains an optional deprecation message.
	TopicParameters []TopicParameter // TopicParameters describes the parameters in the topic pattern (e.g., {deviceID}).
	MessageType     any              // MessageType is the Go type of the message being published.
	QoS             QoS              // QoS is the quality of service level for this publication.
	Retained        bool             // Retained indicates whether the message should be retained by the broker.
	Examples        map[string]any   // Examples contains named examples of messages that can be published.
}

// SubscriptionSpec describes an MQTT subscription operation.
type SubscriptionSpec struct {
	OperationID     string              // OperationID is a unique identifier for this subscription operation (e.g., "subscribeTemperature").
	TopicMQTT       string              // TopicMQTT is the MQTT wildcard format (e.g., devices/+/temperature).
	Summary         string              // Summary is a short description of the subscription.
	Description     string              // Description provides detailed information about the subscription.
	Group           string              // Group is a logical grouping for the subscription (e.g., "Telemetry", "Control").
	Deprecated      string              // Deprecated contains an optional deprecation message.
	TopicParameters []TopicParameter    // TopicParameters describes the parameters in the topic pattern (e.g., {deviceID}).
	MessageType     any                 // Expected Go type of messages received on this subscription.
	Handler         paho.MessageHandler // Handler is the function that will be called when a message is received.
	QoS             QoS                 // QoS is the quality of service level for this subscription.
	Examples        map[string]any      // Examples contains named examples of messages that may be received.
}
