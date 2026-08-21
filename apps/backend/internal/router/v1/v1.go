package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/repositories"
)

func RegisterV1Routes(
	router gin.IRouter,
	handlers *handlers.Handlers,
	middleware *middleware.Middlewares,
	sessionRepo *repositories.SessionRepository,
	jwtSecret []byte,
) {
	// Register users routes
	registerUserRoutes(router, handlers.User, sessionRepo, jwtSecret)

	// Register Auth routes
	registerAuthRoutes(router, handlers.Auth, sessionRepo, jwtSecret)

	// Register Expenses routes
	registerExpenseRoutes(router, handlers.Expense, sessionRepo, jwtSecret)

	//Register Expense Analytics routes
	registerExpenseAnalyticsRouted(router, handlers.ExpenseAnalytics, sessionRepo, jwtSecret)
}
