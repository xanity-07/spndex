package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/repositories"
)

func registerExpenseAnalyticsRouted(r gin.IRouter, h *handlers.ExpenseAnalyticsHandler, sessionRepo *repositories.SessionRepository, jwtSecret []byte) {
	expenseAnalytics := r.Group("/expenses/analytics")

	protected := expenseAnalytics.Group("")
	protected.Use(middleware.RequireAuth(sessionRepo, jwtSecret))
	protected.GET("/category-totals", h.GetExpensesByCategory())
	protected.GET("/monthly-expenses", h.GetMonthlyExpenses())
	protected.GET("/dashboard", h.GetDashboardStats())
	protected.GET("/spending-trends", h.GetSpendingTrends())
}
