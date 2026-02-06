package cloudapi

type CustomID string
type AliasOfCustomID CustomID

// HealthResponse is the response to a health check request.
type HealthResponse struct {
	// Status of the database connection
	Database bool `json:"database"`
}

// PingResponse is the response to a ping request.
type PingResponse struct {
	// Human-readable message
	Message string `json:"message"`
	// Status of the ping
	Status   PingStatus `json:"status"`
	Metadata *string    `json:"metadata,omitempty"`
}

// PingStatus represents the status of a ping request.
type PingStatus string

const (
	// PingStatusOK means the ping was successful.
	PingStatusOK PingStatus = "OK"
	// PingStatusError means there was an error with the ping.
	PingStatusError PingStatus = "ERROR"
)
