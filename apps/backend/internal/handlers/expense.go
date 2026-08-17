package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type ExpenseHandler struct {
	Handler
	expenseService *service.ExpenseService
}

func NewExpenseHandler(s *server.Server, expenseService *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		Handler:        NewHandler(s),
		expenseService: expenseService,
	}
}

func (h *ExpenseHandler) CreateExpense() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *expense.CreateExpensePayload) (*expense.Expense, error) {
			userID, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			id, ok := userID.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			return h.expenseService.CreateExpense(c, id, payload)
		},
		http.StatusCreated,
		func() *expense.CreateExpensePayload {
			return &expense.CreateExpensePayload{}
		},
		enums.BindingJSON,
	)
}

func (h *ExpenseHandler) GetExpenses() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, query *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error) {
			id, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			userID, ok := id.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			return h.expenseService.GetExpenses(c, userID, query)
		},
		http.StatusOK,
		func() *expense.GetExpensesQuery {
			return &expense.GetExpensesQuery{}
		},
		enums.BindingQuery,
	)
}

func (h *ExpenseHandler) GetExpenseByID() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *expense.GetExpenseByID) (*expense.Expense, error) {
			id, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			userID, ok := id.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			val := c.Param("id")

			payload.ID = uuid.MustParse(val).String()
			return h.expenseService.GetExpenseByID(c, userID, payload)
		},
		http.StatusOK,
		func() *expense.GetExpenseByID {
			return &expense.GetExpenseByID{}
		},
		enums.BindingURI,
	)
}

func (h *ExpenseHandler) UpdateExpense() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, payload *expense.UpdateExpense) (*expense.Expense, error) {
		expID := c.Param("id")

		val, ok := c.Get(middleware.UserIDKey)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		userID, ok := val.(uuid.UUID)
		if !ok {
			return nil, errs.NewUnauthorizedError("unauthorized", false)
		}

		val = c.Param("id")

		return h.expenseService.UpdateExpense(c, userID, expID, payload)
	},
		http.StatusOK,
		func() *expense.UpdateExpense {
			return &expense.UpdateExpense{}
		},
		enums.BindingJSON,
	)
}

func (h *ExpenseHandler) DeleteExpense() gin.HandlerFunc {
	return HandleNoContent(h.Handler, func(c *gin.Context, payload *expense.DeleteExpense) error {
		id, ok := c.Get(middleware.UserIDKey)
		if !ok {
			return errs.NewUnauthorizedError("unauthorized", false)
		}

		userID, ok := id.(uuid.UUID)
		if !ok {
			return errs.NewUnauthorizedError("unauthorized", false)
		}

		val := c.Param("id")

		payload.ID = uuid.MustParse(val).String()

		return h.expenseService.DeleteExpense(c, userID, payload)
	},
		http.StatusNoContent,
		&expense.DeleteExpense{},
		enums.BindingURI,
	)
}
