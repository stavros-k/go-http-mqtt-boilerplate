package types

type CustomID string
type AliasOfCustomID CustomID

// HealthResponse is the response to a health check request.
type HealthResponse struct {
	// Status of the database connection
	Database bool `json:"database"`
}
