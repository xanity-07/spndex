package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/xanity-07/spndex/internal/errs"

	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/sqlerr"
)

type GlobalMiddlewares struct {
	server *server.Server
}

func NewGlobalMiddlewares(s *server.Server) *GlobalMiddlewares {
	return &GlobalMiddlewares{
		server: s,
	}
}

func (global *GlobalMiddlewares) CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: global.server.Config.Server.CORSAllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE"},
	})
}

func (global *GlobalMiddlewares) RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Execute the rest of the middlewares/handlers
		c.Next()

		latency := time.Since(start)

		statusCode := c.Writer.Status()
		ginErr := c.Errors.Last()

		// If handlers put an HTTP error into c.Errors,
		// determine the status from that error.
		if ginErr != nil {
			var appErr *errs.AppError

			if errors.As(ginErr.Err, &appErr) {
				statusCode = appErr.Status
			}
		}

		logger := GetLogger(c)

		var event *zerolog.Event

		switch {
		case statusCode >= 500:
			event = logger.Error()

			if ginErr != nil {
				event = event.Err(ginErr.Err)
			}

		case statusCode >= 400:
			event = logger.Warn()

			if ginErr != nil {
				event = event.Err(ginErr.Err)
			}

		default:
			event = logger.Info()
		}

		// Add request ID if available
		if requestID := GetRequestID(c); requestID != "" {
			event = event.Str("request_id", requestID)
		}

		// Add user context if available
		if userID := GetUserID(c); userID != "" {
			event = event.Str("user_id", userID)
		}

		event.
			Dur("duration", latency).
			Int("status", statusCode).
			Str("method", c.Request.Method).
			Str("uri", c.Request.RequestURI).
			Str("host", c.Request.Host).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Msg("API")
	}
}

func (global *GlobalMiddlewares) Recover() gin.HandlerFunc {
	return gin.CustomRecovery(
		func(c *gin.Context, recovered any) {
			logger := GetLogger(c)

			logger.Error().
				Any("panic", recovered).
				Msg("panic recovered")

			c.AbortWithStatusJSON(http.StatusInternalServerError, &errs.AppError{
				Code:     "INTERNAL_SERVER_ERROR",
				Message:  "Internal server error",
				Status:   http.StatusInternalServerError,
				Override: false,
				Action:   &errs.Action{},
				Errors: []errs.FieldError{
					{
						Field: "recovery",
						Error: fmt.Sprint(recovered),
					},
				},
			})
		},
	)
}

func (global *GlobalMiddlewares) Secure() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")           // Clickjacking via iframes
		c.Header("X-Content-Type-Options", "nosniff") // MIME-type sniffing attacks
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")                  // XSS and data injection
		c.Header("Referrer-Policy", "strict-origin")                               // Leaking sensitive URL parameters to third parties
		c.Header("Permissions-Policy", "geolocation=(), camera=(), microphone=()") // Unauthorized use of browser APIs (camera, mic, etc.)
		if c.Request.URL.Path == "/docs" {
			c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; font-src 'self' https://cdn.jsdelivr.net https://fonts.scalar.com")
		} else {
			c.Header("Content-Security-Policy", "default-src 'self'") //	Protocol downgrade and cookie hijacking
		}
		c.Next()

	}
}

func (global *GlobalMiddlewares) GlobalErrorHandler(err error, c *gin.Context) {
	// First try to handle database errors and conver them to appropriate HTTP errors
	originalError := err

	// Try to handle well known database errors
	// Only do this for errors that habent already been converted to AppError
	var appErr *errs.AppError
	if errors.As(err, &appErr) {
		err = errs.NewNotFoundError("Route not found", false, nil)
	} else {
		// Here we call our sqlerr handler that will convert the database errors
		// to appropriate application errors
		err = sqlerr.HandleError(err)
	}

	// Now process the possibly converted error
	var status int
	var code string
	var message string
	var fieldErrors []errs.FieldError
	var action *errs.Action

	switch {
	case errors.As(err, &appErr):
		status = appErr.Status
		code = appErr.Code
		message = appErr.Message
		fieldErrors = appErr.Errors
		action = appErr.Action
	default:
		status = http.StatusInternalServerError
		code = errs.MakeUpperCaseWithUnderscores(http.StatusText(http.StatusInternalServerError))
		message = http.StatusText(http.StatusInternalServerError)
	}

	// Log the original error to help with debugging
	// Use enhanced logger from context which already includes request_id, method, path, ip, user context, and trace context
	logger := *GetLogger(c)

	logger.Error().Stack().
		Err(originalError).
		Int("status", status).
		Str("error_code", code).
		Msg(message)

	if !c.Writer.Written() {
		c.JSON(status, errs.AppError{
			Action:   action,
			Code:     code,
			Message:  message,
			Errors:   fieldErrors,
			Status:   status,
			Override: appErr != nil && appErr.Override,
		})
	}

}
