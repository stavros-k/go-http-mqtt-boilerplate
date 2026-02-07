package apicommon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"http-mqtt-boilerplate/backend/internal/shared/types"
	"http-mqtt-boilerplate/backend/pkg/router"
	"http-mqtt-boilerplate/backend/pkg/utils"
)

const (
	MaxBodySize     = 1048576 // 1MB
	MaxBodyText     = "1MB"
	RequestIDHeader = "X-Request-ID"

	ReadHeaderTimeout = 5 * time.Second
	ReadTimeout       = 30 * time.Second
	WriteTimeout      = 30 * time.Second
	IdleTimeout       = 120 * time.Second
	ShutdownTimeout   = 30 * time.Second
)

const zeroUUID = "00000000-0000-0000-0000-000000000000"

type HTTPServer struct {
	l      *slog.Logger
	server *http.Server
}

func NewHTTPServer(l *slog.Logger, addr string, handler http.Handler) *HTTPServer {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: ReadHeaderTimeout,
		ReadTimeout:       ReadTimeout,
		WriteTimeout:      WriteTimeout,
		IdleTimeout:       IdleTimeout,
	}
	srv.SetKeepAlivesEnabled(true)

	return &HTTPServer{
		l:      l.With(slog.String("component", "http-server")),
		server: srv,
	}
}

func (s *HTTPServer) StartOnBackground(cancel context.CancelFunc) {
	go func() {
		s.l.Info("starting", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.l.Error("failed", utils.ErrAttr(err))
			cancel()
		}
	}()
}

func (s *HTTPServer) ShutdownWithDefaultTimeout() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// MiddlewareHandler holds the logger for middleware.
type MiddlewareHandler struct {
	l *slog.Logger
}

// NewMiddlewareHandler creates a new middleware handler.
func NewMiddlewareHandler(l *slog.Logger) *MiddlewareHandler {
	return &MiddlewareHandler{l: l}
}

// HandlerFunc is a HTTP handler that can return an error.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// NewError creates a simple error response.
func NewError(statusCode int, message string) *types.ErrorResponse {
	return &types.ErrorResponse{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewValidationError creates a validation error with field-level details.
func NewValidationError(fieldErrors map[string]string) *types.ErrorResponse {
	return &types.ErrorResponse{
		StatusCode: http.StatusBadRequest,
		Message:    "Validation failed",
		Errors:     fieldErrors,
	}
}

// ErrorHandler wraps handlers with error handling.
func ErrorHandler(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := GetLogger(r.Context())
		requestID := GetRequestID(r.Context())

		err := fn(w, r)
		if err == nil {
			return
		}

		// This is an expected HTTP error, we return the actual error to the client
		var httpErr *types.ErrorResponse
		if errors.As(err, &httpErr) {
			httpErr.RequestID = requestID
			l.Warn("handler returned HTTP error", "status", httpErr.StatusCode, "message", httpErr.Message)
			RespondJSON(w, r, httpErr.StatusCode, httpErr)

			return
		}

		// Internal errors get logged with full context, but we return a generic message to the client
		l.Error("internal error", utils.ErrAttr(err))
		RespondJSON(w, r, http.StatusInternalServerError, &types.ErrorResponse{
			RequestID: requestID,
			Message:   "Internal Server Error",
		})
	}
}

// RespondJSON sends a JSON response with given status code
// If data is nil, only headers are sent
// In case of JSON encoding error, it is logged but not returned to client
// but the status code is sent already.
func RespondJSON(w http.ResponseWriter, r *http.Request, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data == nil {
		return
	}

	l := GetLogger(r.Context())
	if err := utils.ToJSONStream(w, data); err != nil {
		// Note that if this fails header has already been written
		// There's not much we can do at this point
		l.Error("failed to encode JSON response", utils.ErrAttr(err))
	}
}

// DecodeJSON decodes JSON from request body with error handling.
//
//nolint:ireturn // Generic functions must return type parameter T
func DecodeJSON[T any](r *http.Request) (T, error) {
	var zero T

	r.Body = http.MaxBytesReader(nil, r.Body, MaxBodySize)

	res, err := utils.FromJSONStream[T](r.Body)
	if err != nil {
		// FIXME: on Go 1.26 use errors.AsType[...]()
		var (
			syntaxError        *json.SyntaxError
			unmarshalTypeError *json.UnmarshalTypeError
			maxBytesError      *http.MaxBytesError
			extraDataError     *utils.ExtraDataAfterJSONError
		)

		switch {
		case errors.As(err, &syntaxError):
			return zero, NewError(http.StatusBadRequest, fmt.Sprintf("Invalid JSON syntax at position %d", syntaxError.Offset))

		case errors.As(err, &unmarshalTypeError):
			return zero, NewError(http.StatusBadRequest, fmt.Sprintf("Invalid type for field '%s'", unmarshalTypeError.Field))

		case errors.Is(err, io.EOF):
			return zero, NewError(http.StatusBadRequest, "Request body is empty")

		case errors.Is(err, io.ErrUnexpectedEOF):
			return zero, NewError(http.StatusBadRequest, "Malformed JSON")

		case errors.As(err, &maxBytesError):
			return zero, NewError(http.StatusRequestEntityTooLarge, "Request body too large (max "+MaxBodyText+")")

		case errors.As(err, &extraDataError):
			return zero, NewError(http.StatusBadRequest, "Request body contains multiple JSON objects")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			// json package formats this as: json: unknown field "fieldname"
			return zero, NewError(http.StatusBadRequest, err.Error())

		default:
			return zero, NewError(http.StatusBadRequest, "Invalid JSON payload")
		}
	}

	return res, nil
}

// GenerateResponses adds standard error responses to the given responses map.
func GenerateResponses(responses map[int]router.ResponseSpec) map[int]router.ResponseSpec {
	if _, exists := responses[http.StatusRequestEntityTooLarge]; !exists {
		responses[http.StatusRequestEntityTooLarge] = router.ResponseSpec{
			Description: "Request entity too large",
			Type:        types.ErrorResponse{},
			Examples: map[string]any{
				"Request Entity Too Large": types.ErrorResponse{
					RequestID: zeroUUID,
					Message:   "Request body too large (max " + MaxBodyText + ")",
				},
			},
		}
	}

	if _, exists := responses[http.StatusInternalServerError]; !exists {
		responses[http.StatusInternalServerError] = router.ResponseSpec{
			Description: "Internal Server Error",
			Type:        types.ErrorResponse{},
			Examples: map[string]any{
				"Internal Server Error": types.ErrorResponse{
					RequestID: zeroUUID,
					Message:   "Internal Server Error",
				},
			},
		}
	}

	return responses
}
