package apitypes

import (
	"time"

	"http-mqtt-boilerplate/backend/pkg/types"
)

type CustomID string
type AliasOfCustomID CustomID

// ErrorResponse is the unified error response type.
// It supports both simple errors (just message) and validation errors (message + field errors).
//
//nolint:errname // ErrorResponse is an API response type, not a traditional error
type ErrorResponse struct {
	// HTTP status code (internal only, not sent to client)
	StatusCode int `json:"-"`
	// Request ID for tracking
	RequestID string `json:"requestID"`
	// High-level error message
	Message string `json:"message"`
	// Field-level validation errors
	Errors map[string]string `json:"errors,omitempty"`
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

// AddError adds a field-level error (builder pattern).
func (e *ErrorResponse) AddError(field, message string) *ErrorResponse {
	if e.Errors == nil {
		e.Errors = make(map[string]string)
	}

	e.Errors[field] = message

	return e
}

// PingResponse is the response to a ping request.
type PingResponse struct {
	// Human-readable message
	Message string `json:"message"`
	// Status of the ping
	Status   PingStatus `json:"status"`
	Metadata *string    `json:"metadata,omitempty"`
	// Deprecated: use Test instead
	Test PingStatus `json:"test,omitempty"`
}

// PingStatus represents the status of a ping request.
type PingStatus string

const (
	// PingStatusOK means the ping was successful.
	PingStatusOK PingStatus = "OK"
	// PingStatusError means there was an error with the ping.
	PingStatusError PingStatus = "ERROR"
)

// CreateUserRequest is the request to create a new user.
type CreateUserRequest struct {
	// Username to create
	Username string `json:"username"`
	// Password to create
	Password string `json:"password"`
}

// CreateUserResponse is the response to a create user request.
type CreateUserResponse struct {
	// ID of the created user
	UserID string `json:"userID"`
	// Creation timestamp
	CreatedAt time.Time `json:"createdAt"`
	// URL to the user
	URL *types.URL `json:"url"`
}

// GetTeamRequest is the request to get a team.
type GetTeamRequest struct {
	// ID of the team to get
	TeamID string `json:"teamID"`
}

// User represents a user in the system.
type User struct {
	// ID of the user
	UserID string `json:"userID"`
	// Name of the user
	//
	// Deprecated: Use UserNameV2 instead.
	Name string `json:"name"`
}

// GetTeamResponse is the response to a get team request.
//
// Deprecated: Use GetTeamResponseV2 instead.
type GetTeamResponse struct {
	// ID of the team
	TeamID string `json:"teamID"`
	// Users in the team
	Users []User `json:"users"`
}

// CreateTeamRequest is the request to create a new team.
type CreateTeamRequest struct {
	// Name of the team to create
	Name string `json:"name"`
}

// TemperatureReading represents a temperature sensor reading from an IoT device.
type TemperatureReading struct {
	// DeviceID is the unique identifier of the device sending the reading
	DeviceID string `json:"deviceID"`
	// Temperature is the measured temperature value
	Temperature float64 `json:"temperature"`
	// Unit is the temperature unit (e.g., "celsius", "fahrenheit")
	Unit string `json:"unit"`
	// Timestamp is when the reading was taken
	Timestamp time.Time `json:"timestamp"`
}

// DeviceCommand represents a command sent to an IoT device.
type DeviceCommand struct {
	// DeviceID is the unique identifier of the target device
	DeviceID string `json:"deviceID"`
	// Command is the command to execute (e.g., "restart", "shutdown", "update_config")
	Command string `json:"command"`
	// Parameters contains optional command parameters
	Parameters map[string]string `json:"parameters,omitempty"`
}

// DeviceStatus represents the status of an IoT device.
type DeviceStatus struct {
	// DeviceID is the unique identifier of the device
	DeviceID string `json:"deviceID"`
	// Status is the current status (e.g., "online", "offline", "error")
	Status string `json:"status"`
	// Uptime is how long the device has been running in seconds
	Uptime int64 `json:"uptime"`
	// Timestamp is when the status was reported
	Timestamp time.Time `json:"timestamp"`
}

// SensorTelemetry represents generic sensor data from an IoT device.
type SensorTelemetry struct {
	// DeviceID is the unique identifier of the device
	DeviceID string `json:"deviceID"`
	// SensorType is the type of sensor (e.g., "temperature", "humidity", "pressure")
	SensorType string `json:"sensorType"`
	// Value is the sensor reading value
	Value float64 `json:"value"`
	// Unit is the unit of measurement
	Unit string `json:"unit"`
	// Timestamp is when the reading was taken
	Timestamp time.Time `json:"timestamp"`
	// Quality is the quality of the reading (0-100)
	Quality int `json:"quality"`
}
