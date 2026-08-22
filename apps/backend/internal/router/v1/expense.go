package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/repositories"
)

func registerExpenseRoutes(r gin.IRouter, h *handlers.ExpenseHandler, sessionRepo *repositories.SessionRepository, jwtSecret []byte) {
	expenses := r.Group("/expenses")

	protected := expenses.Group("")
	protected.Use(middleware.RequireAuth(sessionRepo, jwtSecret))
	protected.POST("", h.CreateExpense())
	protected.GET("", h.GetExpenses())
	protected.GET("/:id", h.GetExpenseByID())
	protected.PATCH("/:id", h.UpdateExpense())
	protected.DELETE("/:id", h.DeleteExpense())
}
