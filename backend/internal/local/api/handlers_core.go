package api

import (
	"net/http"

	localtypes "http-mqtt-boilerplate/backend/internal/local/api/types"
	apitypes "http-mqtt-boilerplate/backend/internal/shared/api"
	sharedtypes "http-mqtt-boilerplate/backend/internal/shared/types"
	"http-mqtt-boilerplate/backend/pkg/router"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) error {
	apitypes.RespondJSON(w, r, http.StatusOK, sharedtypes.PingResponse{
		Message: "Pong", Status: sharedtypes.PingStatusOK,
	})

	return nil
}

func (h *Handler) RegisterPing(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "ping",
		Summary:     "Ping the server",
		Description: "Check if the server is alive",
		Group:       CoreGroup,
		RequestType: nil,
		Handler:     apitypes.ErrorHandler(h.Ping),
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        sharedtypes.PingResponse{},
				Examples: map[string]any{
					"Success": sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK},
				},
			},
		}),
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) error {
	status := h.svc.Core.Health(r.Context())
	resp := localtypes.HealthResponse{
		Database: status.Database,
		MQTT:     status.MQTT,
	}

	code := http.StatusOK
	if !status.Database || !status.MQTT {
		code = http.StatusServiceUnavailable
	}

	apitypes.RespondJSON(w, r, code, resp)

	return nil
}

func (h *Handler) RegisterHealth(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "health",
		Summary:     "Check server health",
		Description: "Check if the server is healthy",
		Group:       CoreGroup,
		RequestType: nil,
		Handler:     apitypes.ErrorHandler(h.Health),
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful health response",
				Type:        localtypes.HealthResponse{},
				Examples: map[string]any{
					"Success": localtypes.HealthResponse{Database: true, MQTT: true},
				},
			},
			503: {
				Description: "Server unavailable",
				Type:        localtypes.HealthResponse{},
				Examples: map[string]any{
					"Database Unavailable": localtypes.HealthResponse{Database: false, MQTT: true},
					"MQTT Unavailable":     localtypes.HealthResponse{Database: true, MQTT: false},
					"Both Unavailable":     localtypes.HealthResponse{Database: false, MQTT: false},
				},
			},
			500: {
				Description: "Internal server error",
				Type:        sharedtypes.ErrorResponse{},
			},
		}),
	})
}
