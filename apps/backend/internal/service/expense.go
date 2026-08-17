package service

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseService struct {
	server      *server.Server
	expenseRepo *repositories.ExpenseRepository
	userRepo    *repositories.UserRepository
}

func NewExpenseService(s *server.Server, expenseRepo *repositories.ExpenseRepository, userRepo *repositories.UserRepository) *ExpenseService {
	return &ExpenseService{
		server:      s,
		expenseRepo: expenseRepo,
		userRepo:    userRepo,
	}
}

func (s *ExpenseService) CreateExpense(ctx *gin.Context, userID uuid.UUID, payload *expense.CreateExpensePayload) (*expense.Expense, error) {
	logger := middleware.GetLogger(ctx)

	expense, err := s.expenseRepo.CreateExpense(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create expense")
		return nil, fmt.Errorf("failed to create expense: %w", err)
	}

	// Business event logs
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "expense_created").
		Str("id", expense.ID.String()).
		Str("user_id", expense.UserID).
		Str("category", string(expense.Category)).
		Int("amount", int(expense.Amount)).
		Str("date", expense.Date).
		Msg("Expense created successfully")

	return expense, nil
}

func (s *ExpenseService) GetExpenses(ctx *gin.Context, userID uuid.UUID, query *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error) {
	logger := middleware.GetLogger(ctx)

	expenseList, err := s.expenseRepo.GetExpenses(ctx, userID, query)
	if err != nil {
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("failed to fetch expenses")
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}

	return expenseList, nil
}

func (s *ExpenseService) GetExpenseByID(ctx *gin.Context, userID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error) {
	logger := middleware.GetLogger(ctx)

	exp, err := s.expenseRepo.GetExpenseByID(ctx, userID, payload)
	if err != nil {
		logger.Error().Err(err).Str("id", payload.ID).Str("user_id", userID.String())
		return nil, fmt.Errorf("failed to fetch expense id=%s: %w", payload.ID, err)
	}

	return exp, nil
}

func (s *ExpenseService) UpdateExpense(
	ctx *gin.Context,
	userID uuid.UUID,
	expenseID string,
	payload *expense.UpdateExpense,
) (*expense.Expense, error) {
	logger := middleware.GetLogger(ctx)

	exp, err := s.expenseRepo.GetExpenseByID(ctx, userID, &expense.GetExpenseByID{ID: expenseID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Error().Err(err).Msg("expense not found")
			code := "EXPENSE_NOT_FOUND"
			return nil, errs.NewNotFoundError("expense not found", false, &code)
		}
		logger.Error().Err(err).Msg("failed to fetch expense")
		return nil, fmt.Errorf("failed to fetch expense: %w", err)
	}

	if payload.Category != nil && !payload.Category.Valid() {
		logger.Error().Err(errors.New("invalid category")).Msg("invalid category payload")
		return nil, errs.NewBadRequestError(
			"invalid category must be one of (food transport utilities entertainment healthcare shopping education other)",
			false,
			nil,
			[]errs.FieldError{
				{
					Field: "category",
					Error: fmt.Sprintf(
						"invalid category %s must be one of (food transport utilities entertainment healthcare shopping education other)", *payload.Category,
					),
				},
			},
			nil,
		)
	}

	if payload.Amount != nil && *payload.Amount < 0 {
		logger.Error().Err(errors.New("amount")).Msg("invalid amount")
		return nil, errs.NewBadRequestError(
			"invalid amount must be greater than 0",
			false,
			nil,
			[]errs.FieldError{
				{
					Field: "amount",
					Error: fmt.Sprintf("invalid amount %d must be greater than 0", *payload.Amount),
				},
			},
			nil,
		)
	}

	updatedExpense, err := s.expenseRepo.UpdateExpense(ctx, userID, exp.ID, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to update expense: %w", err)
	}

	// Business event logs
	eventLogger := middleware.GetLogger(ctx)
	eventLogger.Info().
		Str("event", "user_updated").
		Str("id", updatedExpense.ID.String()).
		Str("user_id", updatedExpense.UserID).
		Int("amount", int(updatedExpense.Amount)).
		Str("category", string(updatedExpense.Category)).
		Str("description", *updatedExpense.Description).
		Str("date", updatedExpense.Date).
		Msg("Updated expense successfully")

	return updatedExpense, nil
}

func (s ExpenseService) DeleteExpense(ctx *gin.Context, userID uuid.UUID, payload *expense.DeleteExpense) error {
	user, err := s.userRepo.GetUserByID(ctx, &user.GetUserByIDPayload{ID: payload.ID})
	if err != nil {
		code := "USER_NOT_FOUND"
		return errs.NewNotFoundError("user not found", false, &code)
	}

	err = s.expenseRepo.DeleteExpense(ctx, user.ID, payload)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
