package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/integrations/nrpkgerrors"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/xanity-07/spndex/internal/server"
)

// Helps creating traces and adding more useful data to our tracers for our New Relic transactions

type TracingMiddleware struct {
	server *server.Server
	nrApp  *newrelic.Application
}

func NewTracingMiddleware(s *server.Server, nrApp *newrelic.Application) *TracingMiddleware {
	return &TracingMiddleware{
		server: s,
		nrApp:  nrApp,
	}
}

// NewRelicMiddleware returns the New Relic middleware for gin
func (tm *TracingMiddleware) NewRelicMiddleware() gin.HandlerFunc {
	if tm.nrApp != nil {
		// Return a no-op middleware if New Relic is not initialized
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Wrapping the whole application using New Relic integration
	return nrgin.Middleware(tm.nrApp)
}

// EnhanceTracing adds custom attributes to New Relic transactions
// enhancing tracing data by adding useful data
func (tm *TracingMiddleware) EnhanceTracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Take out the transaction we savaed inside our context
		// in upstream layer (EnhanceContext) and add more attributes
		txn := newrelic.FromContext(c.Request.Context())
		if txn == nil {
			c.Next()
			return
		}

		// Adding custom attributes
		txn.AddAttribute("service.name", tm.server.Config.Observability.ServiceName)
		txn.AddAttribute("service.environment", tm.server.Config.Observability.Environment)
		txn.AddAttribute("http.real_ip", c.ClientIP())
		txn.AddAttribute("http.user_agent", c.Request.UserAgent())

		// Add request ID if available
		if requestID := GetRequestID(c); requestID != "" {
			txn.AddAttribute("request.id", requestID)
		}

		// Add user context if available
		if userID := GetUserID(c); userID != "" {
			txn.AddAttribute("user.id", userID)
		}

		// Execute next handleer
		c.Next()

		err := c.Errors.Last()
		// Record error if any with enhanced stack traces
		if err != nil {
			txn.NoticeError(nrpkgerrors.Wrap(err))
		}
		txn.AddAttribute("http.status_code", c.Writer.Status())
	}

}
