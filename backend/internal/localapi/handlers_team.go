package localapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"http-mqtt-boilerplate/backend/internal/apicommon"
	"http-mqtt-boilerplate/backend/pkg/types/localapi"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/types"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) error {
	teamID := chi.URLParam(r, "teamID")

	apicommon.RespondJSON(w, r, http.StatusOK, localapi.GetTeamResponse{TeamID: teamID, Users: []localapi.User{{UserID: "Asdf"}}})

	return nil
}

func (h *Handler) RegisterGetTeam(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "getTeam",
		Summary:     "Get a team",
		Description: "Get a team by its ID",
		Group:       TeamGroup,
		Deprecated:  "Use GetTeamResponseV2 instead.",
		Handler:     apicommon.ErrorHandler(h.GetTeam),
		RequestType: nil,
		Parameters: map[string]router.ParameterSpec{
			"teamID": {
				In:          "path",
				Description: "ID of the team to get",
				Required:    true,
				Type:        new(string),
			},
		},
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        localapi.PingResponse{},
				Examples: map[string]any{
					"example-1": localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK},
				},
			},
			201: {
				Description: "Successful ping response",
				Type:        localapi.GetTeamResponse{},
				Examples: map[string]any{
					"example-1": localapi.GetTeamResponse{TeamID: "123", Users: []localapi.User{{UserID: "123", Name: "John"}}},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localapi.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localapi.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}},
				},
			},
		}),
	})
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) error {
	apicommon.RespondJSON(w, r, http.StatusOK, localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK})

	return nil
}

func (h *Handler) RegisterCreateTeam(path string, rb *router.RouteBuilder) {
	rb.MustPost(path, router.RouteSpec{
		OperationID: "createTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apicommon.ErrorHandler(h.CreateTeam),
		RequestType: &router.RequestBodySpec{
			Type: localapi.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localapi.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        localapi.PingResponse{},
				Examples: map[string]any{
					"example-1": localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localapi.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localapi.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(types.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) error {
	apicommon.RespondJSON(w, r, http.StatusOK, localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK})

	return nil
}

func (h *Handler) RegisterDeleteTeam(path string, rb *router.RouteBuilder) {
	rb.MustDelete(path, router.RouteSpec{
		OperationID: "deleteTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apicommon.ErrorHandler(h.DeleteTeam),
		RequestType: &router.RequestBodySpec{
			Type: localapi.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localapi.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        localapi.PingResponse{},
				Examples: map[string]any{
					"example-1": localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localapi.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localapi.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(types.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}

func (h *Handler) PutTeam(w http.ResponseWriter, r *http.Request) error {
	apicommon.RespondJSON(w, r, http.StatusOK, localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK})

	return nil
}

func (h *Handler) RegisterPutTeam(path string, rb *router.RouteBuilder) {
	rb.MustPut(path, router.RouteSpec{
		OperationID: "putTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apicommon.ErrorHandler(h.PutTeam),
		RequestType: &router.RequestBodySpec{
			Type: localapi.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localapi.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apicommon.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        localapi.PingResponse{},
				Examples: map[string]any{
					"example-1": localapi.PingResponse{Message: "Pong", Status: localapi.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localapi.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localapi.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(types.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}
