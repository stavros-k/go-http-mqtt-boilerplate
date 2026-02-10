package types

import "time"

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
