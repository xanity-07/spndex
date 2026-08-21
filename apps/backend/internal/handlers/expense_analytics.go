package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type ExpenseAnalyticsHandler struct {
	Handler
	ExpenseAnalytics *service.ExpenseAnalyticsService
}

func NewExpenseAnalyticsHandler(s *server.Server, expenseAnalyticsService *service.ExpenseAnalyticsService) *ExpenseAnalyticsHandler {
	return &ExpenseAnalyticsHandler{
		Handler:          NewHandler(s),
		ExpenseAnalytics: expenseAnalyticsService,
	}
}

func (h *ExpenseAnalyticsHandler) GetExpensesByCategory() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, noReq EmptyRequest) ([]analytic.CategoryTotals, error) {
		id, ok := c.Get(middleware.UserIDKey)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		userID, ok := id.(uuid.UUID)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		return h.ExpenseAnalytics.GetExpensesByCategory(c, userID)
	},
		http.StatusOK,
		func() EmptyRequest {
			return EmptyRequest{}
		},
		enums.BindingURI,
	)
}

func (h *ExpenseAnalyticsHandler) GetMonthlyExpenses() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
			id, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			userID, ok := id.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			param := c.Query("year")
			if param != "" {
				year, err := strconv.Atoi(param)
				if err != nil {
					code := "INVALID_QUERY_YEAR"
					return nil, errs.NewBadRequestError("invalid year", false, &code, []errs.FieldError{
						{
							Field: "query: year",
							Error: "invalid year must be a number",
						},
					}, nil)
				}
				payload.Year = &year
			}

			return h.ExpenseAnalytics.GetMonthlyExpenses(c, userID, payload)
		},
		http.StatusOK,
		func() *analytic.GetMonthlyExpensesPayload {
			return &analytic.GetMonthlyExpensesPayload{}
		},
		enums.BindingQuery,
	)
}

func (h *ExpenseAnalyticsHandler) GetDashboardStats() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, query *analytic.GetDashboardQuery) (*analytic.DashboardStats, error) {
		id, ok := c.Get(middleware.UserIDKey)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		userID, ok := id.(uuid.UUID)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		return h.ExpenseAnalytics.GetDashboardStats(c, userID, query)
	},
		http.StatusOK,
		func() *analytic.GetDashboardQuery {
			return &analytic.GetDashboardQuery{}
		},
		enums.BindingQuery,
	)
}

func (h *ExpenseAnalyticsHandler) GetSpendingTrends() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload EmptyRequest) ([]analytic.MonthlyTotals, error) {
			id, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			userID, ok := id.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			return h.ExpenseAnalytics.GetSpendingTrends(c, userID)
		},
		http.StatusOK,
		func() EmptyRequest {
			return EmptyRequest{}
		},
		enums.BindingQuery,
	)
}
