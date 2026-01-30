package mqtt

import (
	"strings"
	"testing"
)

func TestValidateTopicPattern(t *testing.T) {
	tests := []struct {
		name        string
		topic       string
		expectError bool
		errorMsg    string
	}{
		// Valid patterns
		{
			name:        "simple topic",
			topic:       "devices/temperature",
			expectError: false,
		},
		{
			name:        "topic with one parameter",
			topic:       "devices/{deviceID}/temperature",
			expectError: false,
		},
		{
			name:        "topic with multiple parameters",
			topic:       "devices/{deviceID}/sensors/{sensorID}/temperature",
			expectError: false,
		},
		{
			name:        "parameter with underscore",
			topic:       "devices/{device_id}/temperature",
			expectError: false,
		},
		{
			name:        "parameter with numbers",
			topic:       "devices/{device123}/temperature",
			expectError: false,
		},
		{
			name:        "leading slash",
			topic:       "/devices/temperature",
			expectError: true,
			errorMsg:    "leading slash is not allowed",
		},
		{
			name:        "trailing slash",
			topic:       "devices/temperature/",
			expectError: true,
			errorMsg:    "trailing slash is not allowed",
		},

		// Invalid patterns
		{
			name:        "empty topic",
			topic:       "",
			expectError: true,
			errorMsg:    "topic cannot be empty",
		},
		{
			name:        "multi-level wildcard",
			topic:       "devices/#",
			expectError: true,
			errorMsg:    "multi-level wildcard '#' is not supported",
		},
		{
			name:        "single-level wildcard",
			topic:       "devices/+/temperature",
			expectError: true,
			errorMsg:    "wildcard '+' is not supported",
		},
		{
			name:        "parameter starts with number",
			topic:       "devices/{1device}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter name '1device'",
		},
		{
			name:        "parameter starts with underscore",
			topic:       "devices/{_device}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter name '_device'",
		},
		{
			name:        "parameter with hyphen",
			topic:       "devices/{device-id}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter name 'device-id'",
		},
		{
			name:        "parameter with space",
			topic:       "devices/{device id}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter name 'device id'",
		},
		{
			name:        "incomplete parameter opening brace",
			topic:       "devices/{deviceID/temperature",
			expectError: true,
			errorMsg:    "invalid parameter syntax",
		},
		{
			name:        "incomplete parameter closing brace",
			topic:       "devices/deviceID}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter syntax",
		},
		{
			name:        "empty parameter name",
			topic:       "devices/{}/temperature",
			expectError: true,
			errorMsg:    "invalid parameter name ''",
		},
		{
			name:        "empty segments in middle",
			topic:       "devices//temperature",
			expectError: true,
			errorMsg:    "empty segments are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTopicPattern(tt.topic)
			if tt.expectError {
				if err == nil {
					t.Errorf("validateTopicPattern(%q) expected error containing %q, got nil", tt.topic, tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("validateTopicPattern(%q) error = %q, want error containing %q", tt.topic, err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("validateTopicPattern(%q) unexpected error: %v", tt.topic, err)
				}
			}
		})
	}
}

func TestConvertTopicToMQTT(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no parameters",
			input:    "devices/temperature",
			expected: "devices/temperature",
		},
		{
			name:     "single parameter",
			input:    "devices/{deviceID}/temperature",
			expected: "devices/+/temperature",
		},
		{
			name:     "multiple parameters",
			input:    "devices/{deviceID}/sensors/{sensorID}/temperature",
			expected: "devices/+/sensors/+/temperature",
		},
		{
			name:     "parameter at start",
			input:    "{deviceID}/temperature",
			expected: "+/temperature",
		},
		{
			name:     "parameter at end",
			input:    "devices/{deviceID}",
			expected: "devices/+",
		},
		{
			name:     "all parameters",
			input:    "{type}/{deviceID}/{metric}",
			expected: "+/+/+",
		},
		{
			name:     "empty topic",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertTopicToMQTT(tt.input)
			if result != tt.expected {
				t.Errorf("convertTopicToMQTT(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateQoS(t *testing.T) {
	tests := []struct {
		name        string
		qos         QoS
		expectError bool
	}{
		{
			name:        "QoS 0 - At Most Once",
			qos:         QoSAtMostOnce,
			expectError: false,
		},
		{
			name:        "QoS 1 - At Least Once",
			qos:         QoSAtLeastOnce,
			expectError: false,
		},
		{
			name:        "QoS 2 - Exactly Once",
			qos:         QoSExactlyOnce,
			expectError: false,
		},
		{
			name:        "Invalid QoS 3",
			qos:         3,
			expectError: true,
		},
		{
			name:        "Invalid QoS 255",
			qos:         255,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQoS(tt.qos)
			if tt.expectError && err == nil {
				t.Errorf("validateQoS(%d) expected error, got nil", tt.qos)
			}
			if !tt.expectError && err != nil {
				t.Errorf("validateQoS(%d) unexpected error: %v", tt.qos, err)
			}
		})
	}
}
