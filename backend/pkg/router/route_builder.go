package router

import (
	"errors"
	"fmt"
	"http-mqtt-boilerplate/backend/pkg/generate"
	"http-mqtt-boilerplate/backend/pkg/utils"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

// RouteBuilder is a chi router that collects metadata for OpenAPI generation.
type RouteBuilder struct {
	router    chi.Router
	collector generate.RouteMetadataCollector
	l         *slog.Logger
	prefix    string

	operationIDs map[string]struct{}
}

// NewRouteBuilder creates a new RouteBuilder.
func NewRouteBuilder(l *slog.Logger, collector generate.RouteMetadataCollector) (*RouteBuilder, error) {
	return &RouteBuilder{
		router:       chi.NewRouter(),
		collector:    collector,
		operationIDs: make(map[string]struct{}),
		l:            l.With(slog.String("component", "route-builder")),
	}, nil
}

// Route adds a new route group to the router.
func (rb *RouteBuilder) Route(path string, fn func(rb *RouteBuilder)) {
	oldPrefix := rb.prefix
	rb.prefix += path

	// Isolate sub-router
	rb.router.Group(func(r chi.Router) {
		subRB := &RouteBuilder{
			router:       r,
			collector:    rb.collector,
			operationIDs: rb.operationIDs,
			prefix:       rb.prefix,
			l:            rb.l.With(slog.String("prefix", rb.prefix)),
		}
		fn(subRB)
	})

	rb.prefix = oldPrefix
}

// Use adds middlewares to the router.
func (rb *RouteBuilder) Use(middlewares ...func(http.Handler) http.Handler) *RouteBuilder {
	rb.router.Use(middlewares...)

	return rb
}

// RouteSpec defines a specific route.
type RouteSpec struct {
	OperationID string           // OperationID is the unique identifier for the route
	Handler     http.HandlerFunc // Handler is the function that will handle the route
	Summary     string           // Summary is a short summary of the route
	Description string           // Description is a longer description of the route
	Group       string           // Group is a group name for the route
	Deprecated  string           // Deprecated is a deprecation message for the route

	RequestType *RequestBodySpec     // RequestType is the type of the request body, or nil if no body
	Responses   map[int]ResponseSpec // Responses is a map of status code to response spec

	Parameters map[string]ParameterSpec // Parameters (ie query, path, etc) is a map of parameter name to parameter spec

	// Internal fields
	localPath string // localPath is the path without the prefix
	fullPath  string // fullPath is the full path with the prefix
	method    string // method is the HTTP method (e.g., GET, POST, etc.)
}

type ParameterIn string

const (
	ParameterInPath   ParameterIn = "path"
	ParameterInQuery  ParameterIn = "query"
	ParameterInHeader ParameterIn = "header"
)

// ParameterSpec defines a parameter for a route.
type ParameterSpec struct {
	In          ParameterIn
	Description string
	Required    bool
	Type        any // The Go type - validation comes from interfaces
}

type RequestBodySpec struct {
	Type     any
	Examples map[string]any
}

type ResponseSpec struct {
	Description string
	Type        any
	Examples    map[string]any
}

// Get adds a GET route to the router.
func (rb *RouteBuilder) Get(path string, spec RouteSpec) error {
	spec.method = http.MethodGet

	return rb.add(path, spec)
}

// MustGet adds a GET route to the router and terminates the program if an error occurs.
func (rb *RouteBuilder) MustGet(path string, spec RouteSpec) {
	spec.method = http.MethodGet

	if err := rb.add(path, spec); err != nil {
		rb.l.Error("fatal error", utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Post adds a POST route to the router.
func (rb *RouteBuilder) Post(path string, spec RouteSpec) error {
	spec.method = http.MethodPost

	return rb.add(path, spec)
}

// MustPost adds a POST route to the router and terminates the program if an error occurs.
func (rb *RouteBuilder) MustPost(path string, spec RouteSpec) {
	spec.method = http.MethodPost

	if err := rb.add(path, spec); err != nil {
		rb.l.Error("fatal error", utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Put adds a PUT route to the router.
func (rb *RouteBuilder) Put(path string, spec RouteSpec) error {
	spec.method = http.MethodPut

	return rb.add(path, spec)
}

// MustPut adds a PUT route to the router and terminates the program if an error occurs.
func (rb *RouteBuilder) MustPut(path string, spec RouteSpec) {
	spec.method = http.MethodPut

	if err := rb.add(path, spec); err != nil {
		rb.l.Error("fatal error", utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Patch adds a PATCH route to the router.
func (rb *RouteBuilder) Patch(path string, spec RouteSpec) error {
	spec.method = http.MethodPatch

	return rb.add(path, spec)
}

// MustPatch adds a PATCH route to the router and terminates the program if an error occurs.
func (rb *RouteBuilder) MustPatch(path string, spec RouteSpec) {
	spec.method = http.MethodPatch

	if err := rb.add(path, spec); err != nil {
		rb.l.Error("fatal error", utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Delete adds a DELETE route to the router.
func (rb *RouteBuilder) Delete(path string, spec RouteSpec) error {
	spec.method = http.MethodDelete

	return rb.add(path, spec)
}

// MustDelete adds a DELETE route to the router and terminates the program if an error occurs.
func (rb *RouteBuilder) MustDelete(path string, spec RouteSpec) {
	spec.method = http.MethodDelete

	if err := rb.add(path, spec); err != nil {
		rb.l.Error("fatal error", utils.ErrAttr(err))
		os.Exit(1)
	}
}

// Router returns the underlying chi.Router.
//
//nolint:ireturn // we want to return the specific type chi.Router
func (rb *RouteBuilder) Router() chi.Router {
	return rb.router
}

// add adds a new route to the router and collects metadata.
func (rb *RouteBuilder) add(path string, spec RouteSpec) error {
	sanitizedPath := generate.SanitizePath(path)
	if path != sanitizedPath {
		return fmt.Errorf("invalid path %q; sanitized form would be %q", path, sanitizedPath)
	}

	spec.localPath = path
	cleanPath := rb.prefix + spec.localPath
	cleanPath = generate.SanitizePath(cleanPath)
	spec.fullPath = cleanPath

	if _, exists := rb.operationIDs[spec.OperationID]; exists {
		return fmt.Errorf("operation ID %s already exists", spec.OperationID)
	}

	if err := validateRouteSpec(spec); err != nil {
		return fmt.Errorf("invalid route spec: %w", err)
	}

	// Collect parameters metadata
	parameters, err := generateParameters(spec)
	if err != nil {
		return fmt.Errorf("failed to generate parameters: %w", err)
	}

	// Collect request metadata
	var requestInfo *generate.RequestInfo

	if spec.RequestType != nil {
		if spec.RequestType.Type == nil {
			return errors.New("request type is nil")
		}

		requestInfo = &generate.RequestInfo{
			TypeValue: spec.RequestType.Type,
			Examples:  spec.RequestType.Examples,
		}
	}

	// Collect responses metadata
	responses := make(map[int]generate.ResponseInfo)

	for statusCode, respSpec := range spec.Responses {
		responseInfo := generate.ResponseInfo{
			StatusCode:  statusCode,
			TypeValue:   respSpec.Type,
			Description: respSpec.Description,
			Examples:    respSpec.Examples,
		}

		responses[statusCode] = responseInfo
	}

	// Register route with collector
	if err := rb.collector.RegisterRoute(&generate.RouteInfo{
		OperationID: spec.OperationID,
		Method:      spec.method,
		Path:        spec.fullPath,
		Summary:     spec.Summary,
		Description: spec.Description,
		Group:       spec.Group,
		Deprecated:  spec.Deprecated,
		Request:     requestInfo,
		Parameters:  parameters,
		Responses:   responses,
	}); err != nil {
		return fmt.Errorf("failed to register route: %w", err)
	}

	// Everything is good here. Register the route.

	// Register route with router
	rb.router.Method(spec.method, spec.fullPath, spec.Handler)
	rb.operationIDs[spec.OperationID] = struct{}{}

	rb.l.Info("Registered route", slog.String("method", spec.method), slog.String("path", spec.fullPath), slog.String("operationID", spec.OperationID))

	return nil
}
