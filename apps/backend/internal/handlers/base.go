package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/validation"
)

// Handler provides base functionality for all handlers
type Handler struct {
	server *server.Server
}

// NewHandler creates a new base handler
func NewHandler(s *server.Server) Handler {
	return Handler{
		server: s,
	}
}

// HandlerFunc represents a typed handler function that processes a request and returns a response
type HandlerFunc[Req validation.Validatable, Res any] func(c *gin.Context, payload Req) (Res, error)

// HandlerFuncNoContent represents a typed handler function that processes a request without returning content
type HandlerFuncNoContent[Req validation.Validatable] func(c *gin.Context, payload Req) error

// ResponseHandler defines the interface for handling different response types
type ResponseHandler interface {
	Handle(c *gin.Context, result any)
	GetOperation() string
	AddAttribute(txn *newrelic.Transaction, result any)
}

// JSONResponseHandler handles JSON responses
type JSONResponseHandler struct {
	status int
}

func (h JSONResponseHandler) Handle(c *gin.Context, result any) {
	c.JSON(h.status, result)
}

func (h JSONResponseHandler) GetOperation() string {
	return "handler"
}

func (h JSONResponseHandler) AddAttribute(txn *newrelic.Transaction, result any) {
	// http.status_code is already set by tracing middleware
}

// NoContentResponseHandler handles no-content responses
type NoContentResponseHandler struct {
	status int
}

func (h NoContentResponseHandler) Handle(c *gin.Context, result any) {
	c.Status(http.StatusNoContent)
}

func (h NoContentResponseHandler) GetOperation() string {
	return "handler"
}

func (h NoContentResponseHandler) AddAttribute(txn *newrelic.Transaction, result any) {
	// http.status_code is already set by tracing middleware
}

// handleRequest is the unified handler function that eliminates code duplication
func handleRequest[Req validation.Validatable](
	c *gin.Context,
	payload Req,
	handler func(c *gin.Context, result Req) (any, error),
	responseHandler ResponseHandler,
) {
	start := time.Now()
	method := c.Request.Method
	path := c.Request.URL.Path
	route := path

	// Get New Relic transaction from context
	txn := newrelic.FromContext(c.Request.Context())
	if txn != nil {
		txn.AddAttribute("handler.name", route)
		// http.method and http.route are already set by nrecho middleware
		responseHandler.AddAttribute(txn, nil)
	}

	// Get context-enhanced logger
	logger := middleware.GetLogger(c).With().
		Str("operation", responseHandler.GetOperation()).
		Str("method", method).
		Str("path", path).
		Str("route", route).
		Logger()

	// user.id is already set by tracing middleware

	logger.Info().Msg("handling request")

	// Validation with observability
	validationStart := time.Now()
	if err := validation.BindAndValidate(c, payload); err != nil {
		validationDuration := time.Since(validationStart)

		logger.Error().
			Err(err).
			Dur("validation_duration", validationDuration).
			Msg("request validation failed")

		if txn != nil {
			txn.NoticeError(nrpkgerrors.Wrap(err))
			txn.AddAttribute("validation.status", "failed")

			txn.AddAttribute("validation.duration_ms", validationDuration.Milliseconds())
		}
		return
	}

	validationDuration := time.Since(validationStart)
	if txn != nil {
		txn.AddAttribute("validation.status", "success")
		txn.AddAttribute("validation.duration_ms", validationDuration.Milliseconds())
	}

	logger.Info().
		Dur("validation_duration", validationDuration).
		Msg("request validation success")

	// Execute hancler with observability
	handlerStart := time.Now()
	result, err := handler(c, payload)
	handlerDuration := time.Since(handlerStart)

	if err != nil {
		totalDuration := time.Since(start)

		logger.Error().
			Err(err).
			Dur("handler_duration", handlerDuration).
			Dur("total_duration", totalDuration).
			Msg("handler execution failed")

		if txn != nil {
			txn.NoticeError(nrpkgerrors.Wrap(err))
			txn.AddAttribute("handler.status", "error")
			txn.AddAttribute("handler.duration_ms", handlerDuration.Milliseconds())
			txn.AddAttribute("total.duration_ms", totalDuration.Milliseconds())
		}
		errs.WriteHTTPError(c, err)
		return
	}

	totalDuration := time.Since(start)

	// Record success metrics and tracing
	if txn != nil {
		txn.AddAttribute("handler.status", "success")
		txn.AddAttribute("handler.duration_ms", handlerDuration.Milliseconds())
		txn.AddAttribute("total.duration_ms", totalDuration.Milliseconds())
		responseHandler.AddAttribute(txn, result)
	}

	logger.Info().
		Dur("handler_duration", handlerDuration).
		Dur("validation_duration", validationDuration).
		Dur("total_duration", totalDuration).
		Msg("request completed successfully")

	responseHandler.Handle(c, result)
}

// Handle wraps a handler with validation, error handling, logging, metrics, and tracing
func Handle[Req validation.Validatable, Res any](
	h Handler,
	handler HandlerFunc[Req, Res],
	status int,
	payload Req,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handleRequest(ctx, payload, func(ctx *gin.Context, req Req) (interface{}, error) {
			return handler(ctx, req)
		}, JSONResponseHandler{status: status})
	}
}

// HandleNoContent wraps a handler with validation, error handling, logging, metrics, and tracing for endpoints that don't return content
func HandleNoContent[Req validation.Validatable, Res any](
	h Handler,
	handler HandlerFuncNoContent[Req],
	status int,
	req Req,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		handleRequest(ctx, req, func(ctx *gin.Context, req Req) (interface{}, error) {
			return nil, handler(ctx, req)
		}, NoContentResponseHandler{status: status})
	}
}
