package handlers

import (
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	User    *UserHandlers
}

// pass in service when implemented
func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		User:    NewUserHandlers(s, services.User),
	}
}
