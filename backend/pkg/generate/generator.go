package generate

// RouteMetadataCollector is the interface that RouteBuilder uses to collect route metadata.
type RouteMetadataCollector interface {
	RegisterRoute(route *RouteInfo) error
	Generate() error
}

// MQTTMetadataCollector is the interface that MQTTBuilder uses to collect MQTT operation metadata.
type MQTTMetadataCollector interface {
	RegisterMQTTPublication(pub *MQTTPublicationInfo) error
	RegisterMQTTSubscription(sub *MQTTSubscriptionInfo) error
}

// MetadataCollector is a unified interface that includes both HTTP and MQTT metadata collection.
type MetadataCollector interface {
	RouteMetadataCollector
	MQTTMetadataCollector
}

// NoopCollector is a no-op implementation of MetadataCollector.
type NoopCollector struct{}

func (n *NoopCollector) RegisterRoute(route *RouteInfo) error                     { return nil }
func (n *NoopCollector) RegisterMQTTPublication(pub *MQTTPublicationInfo) error   { return nil }
func (n *NoopCollector) RegisterMQTTSubscription(sub *MQTTSubscriptionInfo) error { return nil }
func (n *NoopCollector) Generate() error                                          { return nil }

// Type kind constants for TypeInfo.
const (
	TypeKindObject     = "object"
	TypeKindStringEnum = "string_enum"
	TypeKindNumberEnum = "number_enum"
	TypeKindAlias      = "alias"
)

// Field type kind constants for FieldType.
const (
	FieldKindPrimitive = "primitive"
	FieldKindArray     = "array"
	FieldKindReference = "reference"
	FieldKindEnum      = "enum"
	FieldKindObject    = "object"
)
