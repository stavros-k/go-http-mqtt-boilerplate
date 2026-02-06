package types

import (
	"time"

	"http-mqtt-boilerplate/backend/pkg/utils"
)

// User represents a user in the system.
type User struct {
	// ID of the user
	UserID string `json:"userID"`
	// Name of the user
	//
	// Deprecated: Use UserNameV2 instead.
	Name string `json:"name"`
}

// GetTeamRequest is the request to get a team.
type GetTeamRequest struct {
	// ID of the team to get
	TeamID string `json:"teamID"`
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
	URL *utils.URL `json:"url"`
}

// HealthResponse is the response to a health check request for local API.
// Local API includes both database and MQTT status.
type HealthResponse struct {
	// Status of the database connection
	Database bool `json:"database"`
	// Status of the MQTT broker connection
	MQTT bool `json:"mqtt"`
}
