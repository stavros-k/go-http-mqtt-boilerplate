package localapi

import (
	"net/http"

	"http-mqtt-boilerplate/backend/internal/shared/api"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/types/common"
	"http-mqtt-boilerplate/backend/pkg/types/localapi"
)

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) error {
	apicommon.RespondJSON(w, r, http.StatusOK, localapi.PingResponse{
		Message: "Pong", Status: localapi.PingStatusOK,
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
				Type:        localapi.PingResponse{},
				Examples: map[string]any{
					"Success": localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK},
				},
			},
		}),
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) error {
	status := h.svc.Core.Health(r.Context())
	resp := localapi.HealthResponse{
		Database: status.Database,
		MQTT:     status.MQTT,
	}

	code := http.StatusOK
	if !status.Database || !status.MQTT {
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
				Type:        localapi.HealthResponse{},
				Examples: map[string]any{
					"Success": localapi.HealthResponse{Database: true, MQTT: true},
				},
			},
			503: {
				Description: "Server unavailable",
				Type:        localapi.HealthResponse{},
				Examples: map[string]any{
					"Database Unavailable": localapi.HealthResponse{Database: false, MQTT: true},
					"MQTT Unavailable":     localapi.HealthResponse{Database: true, MQTT: false},
					"Both Unavailable":     localapi.HealthResponse{Database: false, MQTT: false},
				},
			},
			500: {
				Description: "Internal server error",
				Type:        common.ErrorResponse{},
			},
		}),
	})
}
