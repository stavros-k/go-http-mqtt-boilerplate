package api

import (
	"http-mqtt-boilerplate/backend/pkg/apitypes"
	"http-mqtt-boilerplate/backend/pkg/router"
	"net/http"
)

func (s *Handler) Ping(w http.ResponseWriter, r *http.Request) error {
	if !s.svc.Core.Ping(r.Context()) {
		RespondJSON(w, r, http.StatusInternalServerError, apitypes.PingResponse{
			Message: "Database unreachable", Status: apitypes.PingStatusError,
		})

		return nil
	}

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
			500: {
				Description: "Internal server error",
				Type:        apitypes.PingResponse{},
				Examples: map[string]any{
					"Database unreachable": apitypes.PingResponse{Message: "Database unreachable", Status: apitypes.PingStatusError},
				},
			},
		}),
	})
}
