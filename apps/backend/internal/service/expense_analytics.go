package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseAnalyticsService struct {
	server               *server.Server
	userRepo             *repositories.UserRepository
	expenseAnalyticsRepo *repositories.ExpenseAnalyticsRepository
}

func NewExpenseAnalyticsService(
	s *server.Server,
	expenseAnalyticsRepo *repositories.ExpenseAnalyticsRepository,
	userRepo *repositories.UserRepository,
) *ExpenseAnalyticsService {
	return &ExpenseAnalyticsService{
		server:               s,
		expenseAnalyticsRepo: expenseAnalyticsRepo,
		userRepo:             userRepo,
	}
}

func (s *ExpenseAnalyticsService) GetExpensesByCategory(ctx *gin.Context, userID uuid.UUID) ([]analytic.CategoryTotals, error) {
	logger := middleware.GetLogger(ctx)

	user, err := s.userRepo.GetUserByID(ctx, &user.GetUserByIDPayload{ID: userID.String()})
	if err != nil {
		code := "USER_NOT_FOUND"
		logger.Error().Err(err).Str("user_id", userID.String()).Msg("user not found or deleted")
		return nil, errs.NewNotFoundError("user not found or doesn't exist", false, &code)
	}

	categoryTotal, err := s.expenseAnalyticsRepo.GetExpensesByCategory(ctx, user.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to get category total analytics")
		return nil, fmt.Errorf("failed to get category total analytics: %w", err)
	}

	return categoryTotal, nil
}

func (s *ExpenseAnalyticsService) GetMonthlyExpenses(
	ctx *gin.Context,
	userID uuid.UUID,
	payload *analytic.GetMonthlyExpensesPayload,
) ([]analytic.MonthlyTotals, error) {
	logger := middleware.GetLogger(ctx)

	if payload.Year == nil {
		year := time.Now().Year()
		payload.Year = &year
	}

	if *payload.Year > time.Now().Year() {
		logger.Error().Err(errors.New("invalid year can't be in the future"))
		code := "INVALID_YEAR"
		return nil, errs.NewBadRequestError("invalid year", false, &code,
			[]errs.FieldError{
				{
					Field: "query: year",
					Error: "year can't be in the future",
				},
			},
			nil,
		)
	}

	monthlyTotals, err := s.expenseAnalyticsRepo.GetMonthlyExpenses(ctx, userID, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly totals: %w", err)
	}

	return monthlyTotals, nil
}

func (s *ExpenseAnalyticsService) GetDashboardStats(ctx *gin.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*analytic.DashboardStats, error) {
	logger := middleware.GetLogger(ctx)

	foundUser, err := s.userRepo.GetUserByID(ctx, &user.GetUserByIDPayload{ID: userID.String()})
	if err != nil {
		logger.Warn().Msg("User not found or is deleted")
		code := "USER_NOT_FOUND"
		return nil, errs.NewNotFoundError("user not found or is deleted", false, &code)
	}

	expenseCount, err := s.expenseAnalyticsRepo.GetExpenseCount(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch total count of expenses")
		return nil, fmt.Errorf("failed to get expense count: %w", err)
	}

	if expenseCount == 0 {
		logger.Info().Msg("user has no expenses")
		code := "USER_HAS_NO_EXPENSES"
		return nil, errs.NewNotFoundError("user has no expenses", false, &code)
	}

	highestExpense, err := s.expenseAnalyticsRepo.GetHighestExpense(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch highest expense")
		return nil, fmt.Errorf("failed to get highest expense: %w", err)
	}

	lowestExpense, err := s.expenseAnalyticsRepo.GetLowestExpense(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch lowest expense")
		return nil, fmt.Errorf("failed to get lowest expense: %w", err)
	}

	totalExpenses, err := s.expenseAnalyticsRepo.GetTotalExpenses(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch total expenses")
		return nil, fmt.Errorf("failed to get total expenses: %w", err)
	}

	avgExpenseAmount, err := s.expenseAnalyticsRepo.GetAverageExpenseAmount(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch average expenses")
		return nil, fmt.Errorf("failed to get average expense amount: %w", err)
	}

	currentMonthTotal, err := s.expenseAnalyticsRepo.MonthlyTotals(ctx, foundUser.ID, query)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch monthly expenses")
		return nil, fmt.Errorf("failed to get current monthly expenses: %w", err)
	}

	*query.Month = *query.Month - 1
	previousMonthTotao, err := s.expenseAnalyticsRepo.MonthlyTotals(ctx, foundUser.ID, &analytic.GetDashboardQuery{
		Year:  query.Year,
		Month: query.Month,
		Range: query.Range,
	})
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch previous month expenses")
		return nil, fmt.Errorf("failed to get previous monthly expenses: %w", err)
	}

	dashboardStats := &analytic.DashboardStats{
		HighestExpense:            *highestExpense,
		LowestExpense:             *lowestExpense,
		TotalExpensesCents:        totalExpenses,
		ExpenseCount:              expenseCount,
		AverageExpenseAmountCents: avgExpenseAmount,
		CurrentMonthTotalCents:    currentMonthTotal,
		LastMonthTotalCents:       previousMonthTotao,
		MonthlyNetChangeCents:     currentMonthTotal - previousMonthTotao,
	}

	return dashboardStats, nil
}

func (s *ExpenseAnalyticsService) GetSpendingTrends(ctx *gin.Context, userID uuid.UUID) ([]analytic.MonthlyTotals, error) {
	logger := middleware.GetLogger(ctx)

	foundUser, err := s.userRepo.GetUserByID(ctx, &user.GetUserByIDPayload{ID: userID.String()})
	if err != nil {
		logger.Error().Err(err).Msg("failed to find user")
		code := "USER_NOT_FOUND"
		return nil, errs.NewNotFoundError("user not found or deleted", false, &code)
	}

	spendingTrends, err := s.expenseAnalyticsRepo.GetSpendingTrends(ctx, foundUser.ID)
	if err != nil {
		logger.Error().Err(err).Msg("failed to fetch spending trends")
		return nil, fmt.Errorf("failed to fetch spending trends: %w", err)
	}

	return spendingTrends, nil
}
