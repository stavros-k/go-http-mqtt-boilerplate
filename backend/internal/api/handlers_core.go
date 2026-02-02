package api

import (
	"http-mqtt-boilerplate/backend/pkg/apitypes"
	"http-mqtt-boilerplate/backend/pkg/router"
	"net/http"
)

func (s *Handler) Ping(w http.ResponseWriter, r *http.Request) error {
	RespondJSON(w, r, http.StatusOK, apitypes.PingResponse{
		Message: "Pong", Status: apitypes.PingStatusOK,
	})

	return nil
}

func (s *Handler) RegisterPing(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "ping",
		Summary:     "Ping the server",
		Description: "Check if the server is alive",
		Group:       CoreGroup,
		RequestType: nil,
		Handler:     ErrorHandler(s.Ping),
		Responses: GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        apitypes.PingResponse{},
				Examples: map[string]any{
					"Success": apitypes.PingResponse{Message: "Pong", Status: apitypes.PingStatusOK},
				},
			},
		}),
	})
}

func (s *Handler) Health(w http.ResponseWriter, r *http.Request) error {
	status := s.svc.Core.Health(r.Context())
	resp := apitypes.HealthResponse{
		Database: status.Database,
		MQTT:     status.MQTT,
	}
	code := http.StatusOK
	if !status.Database || !status.MQTT {
		code = http.StatusServiceUnavailable
	}

	RespondJSON(w, r, code, resp)

	return nil
}

func (s *Handler) RegisterHealth(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "health",
		Summary:     "Check server health",
		Description: "Check if the server is healthy",
		Group:       CoreGroup,
		RequestType: nil,
		Handler:     ErrorHandler(s.Health),
		Responses: GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful health response",
				Type:        apitypes.HealthResponse{},
				Examples: map[string]any{
					"Success": apitypes.HealthResponse{Database: true, MQTT: true},
				},
			},
			503: {
				Description: "Server unavailable",
				Type:        apitypes.HealthResponse{},
				Examples: map[string]any{
					"Database Unavailable": apitypes.HealthResponse{Database: false, MQTT: true},
					"MQTT Unavailable":     apitypes.HealthResponse{Database: true, MQTT: false},
					"Both Unavailable":     apitypes.HealthResponse{Database: false, MQTT: false},
				},
			},
			500: {
				Description: "Internal server error",
				Type:        apitypes.ErrorResponse{},
			},
		}),
	})
}
