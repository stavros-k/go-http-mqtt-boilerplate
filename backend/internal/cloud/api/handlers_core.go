package cloudapi

import (
	"net/http"

	"http-mqtt-boilerplate/backend/internal/shared/api"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/types/cloudapi"
	"http-mqtt-boilerplate/backend/pkg/types/common"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) error {
	apicommon.RespondJSON(w, r, http.StatusOK, cloudapi.PingResponse{
		Message: "Pong", Status: cloudapi.PingStatusOK,
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
		Handler:     apicommon.ErrorHandler(h.Ping),
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        cloudapi.PingResponse{},
				Examples: map[string]any{
					"Success": cloudapi.PingResponse{Message: "Pong", Status: cloudapi.PingStatusOK},
				},
			},
		}),
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) error {
	status := h.svc.Core.Health(r.Context())
	resp := cloudapi.HealthResponse{
		Database: status.Database,
	}

	code := http.StatusOK
	if !status.Database {
		code = http.StatusServiceUnavailable
	}

	apicommon.RespondJSON(w, r, code, resp)

	return nil
}

func (h *Handler) RegisterHealth(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "health",
		Summary:     "Check server health",
		Description: "Check if the server is healthy",
		Group:       CoreGroup,
		RequestType: nil,
		Handler:     apicommon.ErrorHandler(h.Health),
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful health response",
				Type:        cloudapi.HealthResponse{},
				Examples: map[string]any{
					"Success": cloudapi.HealthResponse{Database: true},
				},
			},
			503: {
				Description: "Server unavailable",
				Type:        cloudapi.HealthResponse{},
				Examples: map[string]any{
					"Database Unavailable": cloudapi.HealthResponse{Database: false},
				},
			},
			500: {
				Description: "Internal server error",
				Type:        common.ErrorResponse{},
			},
		}),
	})
}
