package handlers

import "github.com/xanity-07/spndex/internal/server"

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
}

// pass in service when implemented
func NewHandlers(s *server.Server) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
	}
}
