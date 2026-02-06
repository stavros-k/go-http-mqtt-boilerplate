package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	localtypes "http-mqtt-boilerplate/backend/internal/local/api/types"
	apitypes "http-mqtt-boilerplate/backend/internal/shared/api"
	sharedtypes "http-mqtt-boilerplate/backend/internal/shared/types"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) error {
	teamID := chi.URLParam(r, "teamID")

	apitypes.RespondJSON(w, r, http.StatusOK, localtypes.GetTeamResponse{TeamID: teamID, Users: []localtypes.User{{UserID: "Asdf"}}})

	return nil
}

func (h *Handler) RegisterGetTeam(path string, rb *router.RouteBuilder) {
	rb.MustGet(path, router.RouteSpec{
		OperationID: "getTeam",
		Summary:     "Get a team",
		Description: "Get a team by its ID",
		Group:       TeamGroup,
		Deprecated:  "Use GetTeamResponseV2 instead.",
		Handler:     apitypes.ErrorHandler(h.GetTeam),
		RequestType: nil,
		Parameters: map[string]router.ParameterSpec{
			"teamID": {
				In:          "path",
				Description: "ID of the team to get",
				Required:    true,
				Type:        new(string),
			},
		},
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        sharedtypes.PingResponse{},
				Examples: map[string]any{
					"example-1": sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK},
				},
			},
			201: {
				Description: "Successful ping response",
				Type:        localtypes.GetTeamResponse{},
				Examples: map[string]any{
					"example-1": localtypes.GetTeamResponse{TeamID: "123", Users: []localtypes.User{{UserID: "123", Name: "John"}}},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localtypes.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localtypes.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}},
				},
			},
		}),
	})
}

func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) error {
	apitypes.RespondJSON(w, r, http.StatusOK, sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK})

	return nil
}

func (h *Handler) RegisterCreateTeam(path string, rb *router.RouteBuilder) {
	rb.MustPost(path, router.RouteSpec{
		OperationID: "createTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apitypes.ErrorHandler(h.CreateTeam),
		RequestType: &router.RequestBodySpec{
			Type: localtypes.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localtypes.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        sharedtypes.PingResponse{},
				Examples: map[string]any{
					"example-1": sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localtypes.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localtypes.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(utils.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}

func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) error {
	apitypes.RespondJSON(w, r, http.StatusOK, sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK})

	return nil
}

func (h *Handler) RegisterDeleteTeam(path string, rb *router.RouteBuilder) {
	rb.MustDelete(path, router.RouteSpec{
		OperationID: "deleteTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apitypes.ErrorHandler(h.DeleteTeam),
		RequestType: &router.RequestBodySpec{
			Type: localtypes.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localtypes.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        sharedtypes.PingResponse{},
				Examples: map[string]any{
					"example-1": sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localtypes.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localtypes.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(utils.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}

func (h *Handler) PutTeam(w http.ResponseWriter, r *http.Request) error {
	apitypes.RespondJSON(w, r, http.StatusOK, sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK})

	return nil
}

func (h *Handler) RegisterPutTeam(path string, rb *router.RouteBuilder) {
	rb.MustPut(path, router.RouteSpec{
		OperationID: "putTeam",
		Summary:     "Create a team",
		Description: "Create a team by its name",
		Group:       TeamGroup,
		Handler:     apitypes.ErrorHandler(h.PutTeam),
		RequestType: &router.RequestBodySpec{
			Type: localtypes.CreateTeamRequest{Name: "My Team"},
			Examples: map[string]any{
				"example-1": localtypes.CreateTeamRequest{Name: "My Team"},
			},
		},
		Responses: apitypes.GenerateResponses(map[int]router.ResponseSpec{
			200: {
				Description: "Successful ping response",
				Type:        sharedtypes.PingResponse{},
				Examples: map[string]any{
					"example-1": sharedtypes.PingResponse{Message: "Pong", Status: sharedtypes.PingStatusOK},
				},
			},
			400: {
				Description: "Invalid request",
				Type:        localtypes.CreateUserResponse{},
				Examples: map[string]any{
					"example-1": localtypes.CreateUserResponse{UserID: "123", CreatedAt: time.Time{}, URL: utils.Ptr(utils.MustNewURL("https://localhost:8080/user"))},
				},
			},
		}),
	})
}
