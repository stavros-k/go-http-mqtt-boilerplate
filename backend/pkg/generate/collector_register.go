package generate

// This file handles registration and validation of HTTP routes and MQTT operations.

import (
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

func (g *OpenAPICollector) RegisterRoute(route *RouteInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(route.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if _, exists := g.httpOps[route.OperationID]; exists {
		return fmt.Errorf("duplicate operationID: %s", route.OperationID)
	}

	// Process request body if provided
	// route.Request can be nil (operation has no request body)
	// If route.Request is provided, its TypeValue must be a valid (non-nil) type
	if route.Request != nil {
		if isNilOrNilPointer(route.Request.TypeValue) {
			return fmt.Errorf("request TypeValue must not be nil when Request is provided in route [%s]", route.OperationID)
		}

		typeName, stringifiedExamples, err := g.processHTTPType(route.Request.TypeValue, route.Request.Examples, "request")
		if err != nil {
			return fmt.Errorf("failed to process request type in route [%s]: %w", route.OperationID, err)
		}

		route.Request.TypeName = typeName
		route.Request.ExamplesStringified = stringifiedExamples
	}

	// Process responses (required - every route must have at least one response)
	// Response TypeValue must be a zero-value struct (e.g., MyResponse{})
	// This indicates the type without providing actual data (examples provide the data)
	for statusCode, response := range route.Responses {
		if isNilOrNilPointer(response.TypeValue) {
			return fmt.Errorf("response TypeValue must not be nil in route [%s] for status %d", route.OperationID, statusCode)
		}

		if !isZeroValueStruct(response.TypeValue) {
			return fmt.Errorf("response TypeValue must be zero value struct (e.g., MyResponse{}) in route [%s] for status %d - use Examples for actual data", route.OperationID, statusCode)
		}

		resp := response

		typeName, stringifiedExamples, err := g.processHTTPType(resp.TypeValue, resp.Examples, "response")
		if err != nil {
			return fmt.Errorf("failed to process response type [%s] for status code [%d] in route [%s]: %w", typeName, statusCode, route.OperationID, err)
		}

		resp.TypeName = typeName
		resp.ExamplesStringified = stringifiedExamples
		route.Responses[statusCode] = resp
	}

	for i := range route.Parameters {
		typeName, _, err := g.processHTTPType(route.Parameters[i].TypeValue, nil, "parameter")
		if err != nil {
			return fmt.Errorf("failed to process parameter type in route [%s]: %w", route.OperationID, err)
		}

		route.Parameters[i].TypeName = typeName
	}

	// Add operation keyed by operationID
	g.httpOps[route.OperationID] = route

	return nil
}

func (g *OpenAPICollector) RegisterMQTTPublication(pub *MQTTPublicationInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(pub.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if err := g.validateUniqueOperationID(pub.OperationID); err != nil {
		return err
	}

	// Process message type and examples
	typeName, stringifiedExamples, err := g.processMQTTMessageType(pub.OperationID, pub.TypeValue, pub.Examples, "publication")
	if err != nil {
		return err
	}

	pub.TypeName = typeName
	pub.ExamplesStringified = stringifiedExamples

	for i := range pub.TopicParameters {
		typeName, err := g.processMQTTTopicParameter(pub.OperationID, pub.TopicParameters[i].TypeValue, "publication topic parameter")
		if err != nil {
			return err
		}

		pub.TopicParameters[i].TypeName = typeName
	}

	// Store publication
	g.mqttPublications[pub.OperationID] = pub

	return nil
}

func (g *OpenAPICollector) RegisterMQTTSubscription(sub *MQTTSubscriptionInfo) error {
	// Validate operationID format
	if err := validateOperationIDFormat(sub.OperationID); err != nil {
		return err
	}

	// Validate operationID is unique
	if err := g.validateUniqueOperationID(sub.OperationID); err != nil {
		return err
	}

	// Process message type and examples
	typeName, stringifiedExamples, err := g.processMQTTMessageType(sub.OperationID, sub.TypeValue, sub.Examples, "subscription")
	if err != nil {
		return err
	}

	sub.TypeName = typeName
	sub.ExamplesStringified = stringifiedExamples

	for i := range sub.TopicParameters {
		typeName, err := g.processMQTTTopicParameter(sub.OperationID, sub.TopicParameters[i].TypeValue, "subscription topic parameter")
		if err != nil {
			return err
		}

		sub.TopicParameters[i].TypeName = typeName
	}

	// Store subscription
	g.mqttSubscriptions[sub.OperationID] = sub

	return nil
}

// registerJSONRepresentation registers the JSON representation of a type value.
// It makes sure to only store the largest representation for the type.
func (g *OpenAPICollector) registerJSONRepresentation(value any) error {
	typeName, err := extractTypeNameFromValue(value)
	if err != nil {
		return fmt.Errorf("failed to extract type name: %w", err)
	}

	// Skip primitive types - they don't need JSON representations
	if _, isPrimitive := g.primitiveTypeMapping[typeName]; isPrimitive {
		return nil
	}

	typeInfo, ok := g.types[typeName]
	if !ok {
		return fmt.Errorf("failed to register JSON representation: type %s not found in types map", typeName)
	}

	representation := string(utils.MustToJSONIndent(value))

	// If stored representation is empty or shorter, update it
	if typeInfo.Representations.JSON == "" || len(representation) > len(typeInfo.Representations.JSON) {
		typeInfo.Representations.JSON = representation
	}

	return nil
}

// processHTTPType extracts type name, marks it as HTTP, and registers JSON representations.
// Returns the extracted type name.
func (g *OpenAPICollector) processHTTPType(typeValue any, examples map[string]any, contextMsg string) (string, map[string]string, error) {
	typeName, err := extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract %s type name: %w", contextMsg, err)
	}

	// Mark as used by HTTP (for OpenAPI spec filtering)
	g.markTypeAsHTTP(typeName)

	if err := g.registerJSONRepresentation(typeValue); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for %s type [%s]: %w", contextMsg, typeName, err)
	}

	// Register and stringify examples if provided
	var stringifiedExamples map[string]string

	if examples != nil {
		if err := g.registerExamples(examples); err != nil {
			return "", nil, fmt.Errorf("failed to register JSON representation for %s example: %w", contextMsg, err)
		}

		stringifiedExamples = stringifyExamples(examples)
	}

	return typeName, stringifiedExamples, nil
}

// processMQTTMessageType extracts type information and registers representations for an MQTT message.
// Returns the type name and stringified examples.
func (g *OpenAPICollector) processMQTTMessageType(operationID string, typeValue any, examples map[string]any, messageKind string) (typeName string, stringifiedExamples map[string]string, err error) {
	// Validate type value is not nil
	if isNilOrNilPointer(typeValue) {
		return "", nil, fmt.Errorf("MessageType must not be nil in %s [%s]", messageKind, operationID)
	}

	// Extract type name from zero value using reflection
	typeName, err = extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract message type name: %w", err)
	}

	// Mark as used by MQTT
	g.markTypeAsMQTT(typeName)

	// Register JSON representation
	if err := g.registerJSONRepresentation(typeValue); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for message type: %w", err)
	}

	// Register examples
	if err := g.registerExamples(examples); err != nil {
		return "", nil, fmt.Errorf("failed to register JSON representation for example: %w", err)
	}

	// Stringify examples
	stringifiedExamples = stringifyExamples(examples)

	return typeName, stringifiedExamples, nil
}

// processMQTTTopicParameter extracts type information and registers representations for an MQTT topic parameter.
// Returns the type name.
func (g *OpenAPICollector) processMQTTTopicParameter(operationID string, typeValue any, contextMsg string) (string, error) {
	// Validate type value is not nil
	if isNilOrNilPointer(typeValue) {
		return "", fmt.Errorf("topic parameter TypeValue must not be nil in %s [%s]", contextMsg, operationID)
	}

	// Extract type name from zero value using reflection
	typeName, err := extractTypeNameFromValue(typeValue)
	if err != nil {
		return "", fmt.Errorf("failed to extract topic parameter type name: %w", err)
	}

	// Mark as used by MQTT
	g.markTypeAsMQTT(typeName)

	// Register JSON representation
	if err := g.registerJSONRepresentation(typeValue); err != nil {
		return "", fmt.Errorf("failed to register JSON representation for topic parameter: %w", err)
	}

	return typeName, nil
}

// registerExamples registers JSON representations for a slice of examples.
func (g *OpenAPICollector) registerExamples(examples map[string]any) error {
	for name, ex := range examples {
		if isNilOrNilPointer(ex) {
			return fmt.Errorf("value for example [%s] should not be nil", name)
		}

		if err := g.registerJSONRepresentation(ex); err != nil {
			return err
		}
	}

	return nil
}

// validateUniqueOperationID checks that an operationID is not already used.
func (g *OpenAPICollector) validateUniqueOperationID(operationID string) error {
	if _, exists := g.mqttPublications[operationID]; exists {
		return fmt.Errorf("duplicate operationID (MQTT publication exists): %s", operationID)
	}

	if _, exists := g.mqttSubscriptions[operationID]; exists {
		return fmt.Errorf("duplicate operationID (MQTT subscription exists): %s", operationID)
	}

	if _, exists := g.httpOps[operationID]; exists {
		return fmt.Errorf("duplicate operationID (HTTP operation exists): %s", operationID)
	}

	return nil
}

// ProtocolType represents the type of protocol using a type.
type ProtocolType int

const (
	ProtocolHTTP ProtocolType = iota
	ProtocolMQTT
)

// markTypeAsUsedBy recursively marks a type and all its referenced types as used by a protocol.
func (g *OpenAPICollector) markTypeAsUsedBy(typeName string, protocol ProtocolType) {
	if typeName == "" {
		return
	}

	typeInfo, exists := g.types[typeName]
	if !exists {
		return // Primitive or external type
	}

	// Check if already marked and mark this type
	switch protocol {
	case ProtocolHTTP:
		if typeInfo.UsedByHTTP {
			return
		}

		typeInfo.UsedByHTTP = true
	case ProtocolMQTT:
		if typeInfo.UsedByMQTT {
			return
		}

		typeInfo.UsedByMQTT = true
	}

	// Recursively mark referenced types
	for _, ref := range typeInfo.References {
		g.markTypeAsUsedBy(ref, protocol)
	}
}

// markTypeAsHTTP marks a type and all its referenced types as used by HTTP.
func (g *OpenAPICollector) markTypeAsHTTP(typeName string) {
	g.markTypeAsUsedBy(typeName, ProtocolHTTP)
}

// markTypeAsMQTT marks a type and all its referenced types as used by MQTT.
func (g *OpenAPICollector) markTypeAsMQTT(typeName string) {
	g.markTypeAsUsedBy(typeName, ProtocolMQTT)
}

// stringifyExamples converts examples to stringified JSON.
func stringifyExamples(examples map[string]any) map[string]string {
	stringified := make(map[string]string)
	for name, example := range examples {
		stringified[name] = string(utils.MustToJSONIndent(example))
	}

	return stringified
}

// validateOperationIDFormat checks that an operationID contains only characters a-z, A-Z.
func validateOperationIDFormat(operationID string) error {
	if operationID == "" {
		return errors.New("operationID cannot be empty")
	}

	if !IsASCIILetterString(operationID) {
		return fmt.Errorf("operationID %q contains invalid characters (only characters a-z, A-Z are allowed)", operationID)
	}

	return nil
}
