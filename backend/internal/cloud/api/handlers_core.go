package api

import (
	"net/http"

	cloudtypes "http-mqtt-boilerplate/backend/internal/cloud/api/types"
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
	resp := cloudtypes.HealthResponse{
		Database: status.Database,
	}

	code := http.StatusOK
	if !status.Database {
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
				Type:        cloudtypes.HealthResponse{},
				Examples: map[string]any{
					"Success": cloudtypes.HealthResponse{Database: true},
				},
			},
			503: {
				Description: "Server unavailable",
				Type:        cloudtypes.HealthResponse{},
				Examples: map[string]any{
					"Database Unavailable": cloudtypes.HealthResponse{Database: false},
				},
			},
			500: {
				Description: "Internal server error",
				Type:        sharedtypes.ErrorResponse{},
			},
		}),
	})
}
