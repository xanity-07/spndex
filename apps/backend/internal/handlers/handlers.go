package handlers

import (
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type Handlers struct {
	Health  *HealthHandler
	OpenAPI *OpenAPIHandler
	User    *UserHandlers
	Auth    *AuthHandler
	Expense *ExpenseHandler
}

// pass in service when implemented
func NewHandlers(s *server.Server, services *service.Services) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(s),
		OpenAPI: NewOpenAPIHandler(s),
		User:    NewUserHandlers(s, services.User),
		Auth:    NewAuthHandler(s, services.Auth),
		Expense: NewExpenseHandler(s, services.Expense),
	}
}
