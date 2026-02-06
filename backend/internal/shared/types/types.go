package types

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
}

// PingStatus represents the status of a ping request.
type PingStatus string

const (
	// PingStatusOK means the ping was successful.
	PingStatusOK PingStatus = "OK"
	// PingStatusError means there was an error with the ping.
	PingStatusError PingStatus = "ERROR"
)
