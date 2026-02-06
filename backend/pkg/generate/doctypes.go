package generate

// TypeInfo contains comprehensive metadata about a Go type extracted from guts.
type TypeInfo struct {
	Name            string          `json:"name"`            // Type name (e.g., "PingResponse")
	Kind            string          `json:"kind"`            // "Object", "String Enum", "Array", etc.
	Description     string          `json:"description"`     // Type-level documentation
	Deprecated      string          `json:"deprecated"`      // Deprecation information
	Fields          []FieldInfo     `json:"fields"`          // For object types: fields
	EnumValues      []EnumValue     `json:"enumValues"`      // For enum types: enum constants
	References      []string        `json:"references"`      // Types this type references
	ReferencedBy    []string        `json:"referencedBy"`    // Types that reference this type
	UsedBy          []UsageInfo     `json:"usedBy"`          // Operations/routes that use this type
	Representations Representations `json:"representations"` // JSON, JSON Schema, and TypeScript representations of the type
	UsedByHTTP      bool            `json:"usedByHTTP"`      // Whether this type is used by HTTP operations
	UsedByMQTT      bool            `json:"usedByMQTT"`      // Whether this type is used by MQTT operations
	UnderlyingType  *FieldType      `json:"underlyingType"`  // For alias types: the underlying type being aliased
}

type Representations struct {
	JSON       string `json:"json"`       // JSON representation of the type (zero value example)
	JSONSchema string `json:"jsonSchema"` // JSON Schema representation of the type
	YAMLSchema string `json:"yamlSchema"` // YAML Schema representation of the type (more human-readable)
	Go         string `json:"go"`         // Go representation of the type
	TS         string `json:"ts"`         // TypeScript representation of the type
}

// FieldType represents the structured type information for a field.
type FieldType struct {
	Kind                 string     `json:"kind"`                 // "primitive", "array", "reference", "enum", "object", "unknown"
	Type                 string     `json:"type"`                 // Base type: "string", "User", etc.
	Format               string     `json:"format"`               // OpenAPI format (e.g., "date-time")
	Required             bool       `json:"required"`             // Whether the field is required
	Nullable             bool       `json:"nullable"`             // For nullable types (T | null)
	ItemsType            *FieldType `json:"itemsType"`            // For arrays: type of array elements
	AdditionalProperties *FieldType `json:"additionalProperties"` // For maps: type of map values
	MapKeyType           *FieldType `json:"mapKeyType"`           // For maps: type of map keys
}

// FieldInfo describes a field in a struct (used in high-level API documentation).
type FieldInfo struct {
	Name        string    `json:"name"`        // Field name
	DisplayType string    `json:"displayType"` // Human-readable type string (e.g., "User[]", "string | null")
	TypeInfo    FieldType `json:"typeInfo"`    // Structured type information
	Description string    `json:"description"` // Field documentation
	Deprecated  string    `json:"deprecated"`  // Deprecation information
}

// EnumValue represents an enum constant with its documentation.
type EnumValue struct {
	Value       any    `json:"value"` // string for string enums, int64 for number enums
	Description string `json:"description"`
	Deprecated  string `json:"deprecated"`
}

// UsageInfo tracks where a type is used in operations/routes.
type UsageInfo struct {
	OperationID string `json:"operationID"` // Route operationID
	Role        string `json:"role"`        // "request", "response", "parameter"
}

// RouteInfo contains metadata about a REST route.
type RouteInfo struct {
	OperationID string               `json:"operationID"`
	Method      string               `json:"method"` // HTTP method (GET, POST, etc.)
	Path        string               `json:"path"`   // URL path
	Summary     string               `json:"summary"`
	Description string               `json:"description"`
	Group       string               `json:"group"`
	Deprecated  string               `json:"deprecated"`
	Request     *RequestInfo         `json:"request"`
	Parameters  []ParameterInfo      `json:"parameters"`
	Responses   map[int]ResponseInfo `json:"responses"` // Keyed by status code
}

// RequestInfo describes a request body.
type RequestInfo struct {
	TypeName            string            `json:"type"` // Extracted type name (set by generator)
	TypeValue           any               `json:"-"`    // Zero value of the type (set by route builder)
	Description         string            `json:"description"`
	ExamplesStringified map[string]string `json:"examples"` // Keyed by example name
	Examples            map[string]any    `json:"-"`        // Keyed by example name
}

// ParameterInfo describes a route parameter.
type ParameterInfo struct {
	Name        string `json:"name"`
	In          string `json:"in"`   // "path", "query", "header"
	TypeName    string `json:"type"` // Extracted type name (set by generator)
	TypeValue   any    `json:"-"`    // Zero value of the type (set by route builder)
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ResponseInfo describes a response.
type ResponseInfo struct {
	StatusCode          int               `json:"statusCode"`
	TypeName            string            `json:"type"` // Extracted type name (set by generator), empty for responses without body
	TypeValue           any               `json:"-"`    // Zero value of the type (set by route builder)
	Description         string            `json:"description"`
	ExamplesStringified map[string]string `json:"examples"` // Keyed by example name
	Examples            map[string]any    `json:"-"`        // Keyed by example name
}

// MQTTTopicParameter describes a parameter in an MQTT topic pattern.
type MQTTTopicParameter struct {
	Name        string `json:"name"`        // Parameter name (e.g., "deviceID")
	Description string `json:"description"` // Parameter description
	TypeName    string `json:"type"`        // Parameter type (extracted type name, set by generator)
	TypeValue   any    `json:"-"`           // Zero value of the type (set by mqtt builder)
}

// MQTTPublicationInfo contains metadata about an MQTT publication.
type MQTTPublicationInfo struct {
	OperationID         string               `json:"operationID"`
	Topic               string               `json:"topic"`           // Parameterized topic (e.g., devices/{deviceID}/temperature)
	TopicMQTT           string               `json:"topicMQTT"`       // MQTT wildcard format (e.g., devices/+/temperature)
	TopicParameters     []MQTTTopicParameter `json:"topicParameters"` // Topic parameter descriptions
	Summary             string               `json:"summary"`
	Description         string               `json:"description"`
	Group               string               `json:"group"`
	Deprecated          string               `json:"deprecated"`
	QoS                 byte                 `json:"qos"`
	Retained            bool                 `json:"retained"`
	TypeName            string               `json:"type"`     // Extracted type name (set by generator)
	TypeValue           any                  `json:"-"`        // Zero value of the type (set by mqtt builder)
	ExamplesStringified map[string]string    `json:"examples"` // Keyed by example name
	Examples            map[string]any       `json:"-"`        // Keyed by example name
}

// MQTTSubscriptionInfo contains metadata about an MQTT subscription.
type MQTTSubscriptionInfo struct {
	OperationID         string               `json:"operationID"`
	Topic               string               `json:"topic"`           // Parameterized topic (e.g., devices/{deviceID}/temperature)
	TopicMQTT           string               `json:"topicMQTT"`       // MQTT wildcard format (e.g., devices/+/temperature)
	TopicParameters     []MQTTTopicParameter `json:"topicParameters"` // Topic parameter descriptions
	Summary             string               `json:"summary"`
	Description         string               `json:"description"`
	Group               string               `json:"group"`
	Deprecated          string               `json:"deprecated"`
	QoS                 byte                 `json:"qos"`
	TypeName            string               `json:"type"`     // Extracted type name (set by generator)
	TypeValue           any                  `json:"-"`        // Zero value of the type (set by mqtt builder)
	ExamplesStringified map[string]string    `json:"examples"` // Keyed by example name
	Examples            map[string]any       `json:"-"`        // Keyed by example name
}

// APIDocumentation is the complete API documentation structure.
type APIDocumentation struct {
	Info              APIInfo                          `json:"info"`
	Types             map[string]*TypeInfo             `json:"types"`             // Keyed by type name
	HTTPOperations    map[string]*RouteInfo            `json:"httpOperations"`    // Keyed by operationID
	MQTTPublications  map[string]*MQTTPublicationInfo  `json:"mqttPublications"`  // Keyed by operationID
	MQTTSubscriptions map[string]*MQTTSubscriptionInfo `json:"mqttSubscriptions"` // Keyed by operationID
	Database          Database                         `json:"database"`
	OpenAPISpec       string                           `json:"openapiSpec"` // Stringified OpenAPI YAML specification
}

// APIInfo contains API metadata.
type APIInfo struct {
	Title       string       `json:"title"`
	Version     string       `json:"version"`
	Description string       `json:"description"`
	Servers     []ServerInfo `json:"servers"`
}

// ServerInfo contains server information.
type ServerInfo struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type Database struct {
	Schema string        `json:"schema"`
	Stats  DatabaseStats `json:"stats"`
}

type DatabaseStats struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name        string       `json:"name"`
	Columns     []Column     `json:"columns"`
	ForeignKeys []ForeignKey `json:"foreign_keys,omitempty"`
	Indexes     []Index      `json:"indexes,omitempty"`
}

type Column struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	NotNull    bool    `json:"not_null"`
	Default    *string `json:"default,omitempty"`
	PrimaryKey bool    `json:"primary_key"`
}

type ForeignKey struct {
	From  string `json:"from"`
	Table string `json:"table"`
	To    string `json:"to"`
}

type Index struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
}
