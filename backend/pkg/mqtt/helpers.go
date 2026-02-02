package mqtt

import (
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"strings"
)

// validateTopicPattern validates an MQTT topic pattern with {param} placeholders.
// Valid patterns:
// - Parameters must be in {paramName} format (e.g., devices/{deviceID}/temperature)
// - Parameter names must start with a letter and contain only alphanumeric characters and underscores
// - Multi-level wildcards '#' are NOT supported for explicitness.
func validateTopicPattern(topic string) error {
	if topic == "" {
		return errors.New("topic cannot be empty")
	}

	if strings.HasPrefix(topic, "/") {
		return errors.New("leading slash is not allowed")
	}

	if strings.HasSuffix(topic, "/") {
		return errors.New("trailing slash is not allowed")
	}

	for segment := range strings.SplitSeq(topic, "/") {
		if segment == "" {
			return errors.New("empty segments are not allowed")
		}

		// Check for multi-level wildcard - not allowed
		if strings.Contains(segment, "#") {
			return errors.New("multi-level wildcard '#' is not supported - use explicit parameters {param} instead")
		}

		// Check for single-level wildcard - should use {param} instead
		if strings.Contains(segment, "+") {
			return errors.New("wildcard '+' is not supported - use parameter syntax {param} instead")
		}

		// Check for parameter syntax
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			paramName := segment[1 : len(segment)-1]
			if !generate.IsValidParameterName(paramName) {
				return fmt.Errorf("invalid parameter name '%s' - must start with a letter and contain only alphanumeric characters and underscores", paramName)
			}
		} else if strings.Contains(segment, "{") || strings.Contains(segment, "}") {
			return errors.New("invalid parameter syntax - use {paramName} format")
		}
	}

	return nil
}

// convertTopicToMQTT converts a parameterized topic (devices/{deviceID}/temperature)
// to an MQTT wildcard pattern (devices/+/temperature).
func convertTopicToMQTT(topic string) string {
	segments := strings.Split(topic, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			segments[i] = "+"
		}
	}

	return strings.Join(segments, "/")
}

// validateQoS validates a QoS level.
func validateQoS(qos QoS) error {
	if qos != QoSAtMostOnce && qos != QoSAtLeastOnce && qos != QoSExactlyOnce {
		return errors.New("qos must be 0, 1, or 2")
	}

	return nil
}

func generateParameters(topic string, topicParams []TopicParameter) ([]generate.MQTTTopicParameter, error) {
	var parameters []generate.MQTTTopicParameter
	// Validate path parameters and collect metadata
	params := map[string]struct{}{}
	documentedPathParams := map[string]struct{}{}

	// Extract param names from topic
	for section := range strings.SplitSeq(topic, "/") {
		paramsName, err := generate.ExtractParamName(section)
		if err != nil {
			return nil, fmt.Errorf("invalid topic %s: %w", topic, err)
		}

		for _, paramName := range paramsName {
			params[paramName] = struct{}{}
		}
	}

	// For each documented parameter, validate and collect metadata
	for _, paramSpec := range topicParams {
		if paramSpec.Name == "" {
			return nil, fmt.Errorf("parameter name required for topic %s", topic)
		}

		if paramSpec.Description == "" {
			return nil, fmt.Errorf("parameter Description required for topic %s", topic)
		}

		if paramSpec.Type == nil {
			return nil, fmt.Errorf("parameter Type required for topic %s", topic)
		}

		parameters = append(parameters, generate.MQTTTopicParameter{
			Name:        paramSpec.Name,
			TypeValue:   paramSpec.Type,
			Description: paramSpec.Description,
		})

		if _, exists := params[paramSpec.Name]; !exists {
			return nil, fmt.Errorf("documented parameter %s not found in topic", paramSpec.Name)
		}

		documentedPathParams[paramSpec.Name] = struct{}{}
	}

	// Now go over all discovered path parameters and validate that they are documented
	for name := range params {
		if _, exists := documentedPathParams[name]; !exists {
			return nil, fmt.Errorf("topic parameter %s not documented", name)
		}
	}

	return parameters, nil
}

// validatePublicationSpec validates a publication specification.
func (mb *MQTTBuilder) validatePublicationSpec(spec PublicationSpec) error {
	if spec.OperationID == "" {
		return errors.New("operationID is required")
	}

	if spec.Summary == "" {
		return errors.New("summary is required")
	}

	if spec.Description == "" {
		return errors.New("description is required")
	}

	if spec.Group == "" {
		return errors.New("group is required")
	}

	if spec.MessageType == nil {
		return errors.New("messageType is required")
	}

	if err := validateQoS(spec.QoS); err != nil {
		return err
	}

	return nil
}

// validateSubscriptionSpec validates a subscription specification.
func (mb *MQTTBuilder) validateSubscriptionSpec(spec SubscriptionSpec) error {
	if spec.OperationID == "" {
		return errors.New("operationID is required")
	}

	if spec.Summary == "" {
		return errors.New("summary is required")
	}

	if spec.Description == "" {
		return errors.New("description is required")
	}

	if spec.Group == "" {
		return errors.New("group is required")
	}

	if spec.MessageType == nil {
		return errors.New("messageType is required")
	}

	if spec.Handler == nil {
		return errors.New("handler is required")
	}

	if err := validateQoS(spec.QoS); err != nil {
		return err
	}

	return nil
}
