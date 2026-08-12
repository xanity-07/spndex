package middleware

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/xanity-07/spndex/internal/server"
)

type Middlewares struct {
	Global          *GlobalMiddlewares
	ContextEnhancer *ContextEnhancer
	Tracing         *TracingMiddleware
}

func NewMiddlewares(s *server.Server) *Middlewares {
	// Get New Relic application instance from server
	var nrApp *newrelic.Application
	if s.LoggerService != nil && s.LoggerService.GetApplication() != nil {
		nrApp = s.LoggerService.GetApplication()
	}

	return &Middlewares{
		Global:          NewGlobalMiddlewares(s),
		ContextEnhancer: NewContextEnhancer(s),
		Tracing:         NewTracingMiddleware(s, nrApp),
	}
}
