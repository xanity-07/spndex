package service

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/model/user"
)

type fakeExpenseAnalyticsRepository struct {
	getAverageExpenseAmountFn func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (int64, error)
	getExpenseCountFn         func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (int, error)
	getExpensesByCategoryFn   func(context.Context, uuid.UUID) ([]analytic.CategoryTotals, error)
	getHighestExpenseFn       func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (*expense.Expense, error)
	getLowestExpenseFn        func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (*expense.Expense, error)
	getMonthlyExpensesFn      func(context.Context, uuid.UUID, *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error)
	getSpendingTrendsFn       func(context.Context, uuid.UUID) ([]analytic.MonthlyTotals, error)
	getTotalExpensesFn        func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (int64, error)
	monthlyTotalsFn           func(context.Context, uuid.UUID, *analytic.GetDashboardQuery) (int64, error)
}

func (f *fakeExpenseAnalyticsRepository) GetAverageExpenseAmount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	return f.getAverageExpenseAmountFn(ctx, userID, query)
}

func (f *fakeExpenseAnalyticsRepository) GetExpenseCount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int, error) {
	return f.getExpenseCountFn(ctx, userID, query)
}

func (f *fakeExpenseAnalyticsRepository) GetExpensesByCategory(ctx context.Context, userID uuid.UUID) ([]analytic.CategoryTotals, error) {
	return f.getExpensesByCategoryFn(ctx, userID)
}

func (f *fakeExpenseAnalyticsRepository) GetHighestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error) {
	return f.getHighestExpenseFn(ctx, userID, query)
}

func (f *fakeExpenseAnalyticsRepository) GetLowestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error) {
	return f.getLowestExpenseFn(ctx, userID, query)
}

func (f *fakeExpenseAnalyticsRepository) GetMonthlyExpenses(ctx context.Context, userID uuid.UUID, payload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
	return f.getMonthlyExpensesFn(ctx, userID, payload)
}

func (f *fakeExpenseAnalyticsRepository) GetSpendingTrends(ctx context.Context, userID uuid.UUID) ([]analytic.MonthlyTotals, error) {
	return f.getSpendingTrendsFn(ctx, userID)
}

func (f *fakeExpenseAnalyticsRepository) GetTotalExpenses(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	return f.getTotalExpensesFn(ctx, userID, query)
}

func (f *fakeExpenseAnalyticsRepository) MonthlyTotals(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error) {
	return f.monthlyTotalsFn(ctx, userID, query)
}

func TestExpenseAnalyticsService_GetExpensesByCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	expectedTotals := []analytic.CategoryTotals{
		{
			Category: enums.ExpenseCategory("food"),
			CategoryStats: analytic.CategoryStats{
				TotalCents: 5000,
				Count:      2,
			},
			Percentage: 50,
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				require.Equal(t, userID.String(), payload.ID)

				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpensesByCategoryFn: func(ctx context.Context, gotUserID uuid.UUID) ([]analytic.CategoryTotals, error) {
				require.Equal(t, userID, gotUserID)

				return expectedTotals, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetExpensesByCategory(ctx, userID)

		require.NoError(t, err)
		require.Equal(t, expectedTotals, result)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, errors.New("user not found")
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpensesByCategoryFn: func(ctx context.Context, userID uuid.UUID) ([]analytic.CategoryTotals, error) {
				t.Fatal("GetExpensesByCategory should not be called")
				return nil, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetExpensesByCategory(ctx, userID)

		require.Nil(t, result)
		require.Error(t, err)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "USER_NOT_FOUND", appErr.Code)
	})

	t.Run("analytics repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpensesByCategoryFn: func(ctx context.Context, gotUserID uuid.UUID) ([]analytic.CategoryTotals, error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetExpensesByCategory(ctx, userID)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseAnalyticsService_GetMonthlyExpenses(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	currentYear := time.Now().Year()

	expectedTotals := []analytic.MonthlyTotals{
		{
			Month:      "08-2026",
			TotalCents: 5000,
			Count:      2,
		},
	}

	t.Run("defaults to current year", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		payload := &analytic.GetMonthlyExpensesPayload{}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getMonthlyExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
				require.Equal(t, userID, gotUserID)
				require.NotNil(t, gotPayload.Year)
				require.Equal(t, currentYear, *gotPayload.Year)

				return expectedTotals, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, nil)

		result, err := service.GetMonthlyExpenses(ctx, userID, payload)

		require.NoError(t, err)
		require.Equal(t, expectedTotals, result)
		require.Equal(t, currentYear, *payload.Year)
	})

	t.Run("future year", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		futureYear := currentYear + 1

		payload := &analytic.GetMonthlyExpensesPayload{
			Year: &futureYear,
		}

		repositoryCalled := false

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getMonthlyExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
				repositoryCalled = true
				return nil, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, nil)

		result, err := service.GetMonthlyExpenses(ctx, userID, payload)

		require.Nil(t, result)
		require.Error(t, err)
		require.False(t, repositoryCalled)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "INVALID_YEAR", appErr.Code)
	})

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		year := currentYear - 1

		payload := &analytic.GetMonthlyExpensesPayload{
			Year: &year,
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getMonthlyExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
				require.Equal(t, userID, gotUserID)
				require.Equal(t, payload, gotPayload)

				return expectedTotals, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, nil)

		result, err := service.GetMonthlyExpenses(ctx, userID, payload)

		require.NoError(t, err)
		require.Equal(t, expectedTotals, result)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		year := currentYear

		payload := &analytic.GetMonthlyExpensesPayload{
			Year: &year,
		}

		expectedErr := errors.New("database error")

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getMonthlyExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotPayload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, nil)

		result, err := service.GetMonthlyExpenses(ctx, userID, payload)

		require.Nil(t, result)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseAnalyticsService_GetDashboardStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	year := time.Now().Year()
	month := int(time.Now().Month())

	query := &analytic.GetDashboardQuery{
		Year:  &year,
		Month: &month,
	}

	highestExpense := &expense.Expense{
		UserID:      userID.String(),
		AmountCents: 10000,
	}

	lowestExpense := &expense.Expense{
		UserID:      userID.String(),
		AmountCents: 500,
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				require.Equal(t, userID.String(), payload.ID)

				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		monthlyTotalsCalls := 0

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpenseCountFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int, error) {
				require.Equal(t, userID, gotUserID)
				return 10, nil
			},

			getHighestExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				return highestExpense, nil
			},

			getLowestExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (*expense.Expense, error) {
				require.Equal(t, userID, gotUserID)
				return lowestExpense, nil
			},

			getTotalExpensesFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int64, error) {
				require.Equal(t, userID, gotUserID)
				return 25000, nil
			},

			getAverageExpenseAmountFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int64, error) {
				require.Equal(t, userID, gotUserID)
				return 2500, nil
			},

			monthlyTotalsFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int64, error) {
				require.Equal(t, userID, gotUserID)

				monthlyTotalsCalls++

				if monthlyTotalsCalls == 1 {
					return 8000, nil
				}

				return 5000, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetDashboardStats(ctx, userID, query)

		require.NoError(t, err)
		require.NotNil(t, result)

		require.Equal(t, *highestExpense, result.HighestExpense)
		require.Equal(t, *lowestExpense, result.LowestExpense)
		require.Equal(t, int64(25000), result.TotalExpensesCents)
		require.Equal(t, 10, result.ExpenseCount)
		require.Equal(t, int64(2500), result.AverageExpenseAmountCents)
		require.Equal(t, int64(8000), result.CurrentMonthTotalCents)
		require.Equal(t, int64(5000), result.LastMonthTotalCents)
		require.Equal(t, int64(3000), result.MonthlyNetChangeCents)
		require.Equal(t, 2, monthlyTotalsCalls)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, errors.New("user not found")
			},
		}

		service := NewExpenseAnalyticsService(
			nil,
			&fakeExpenseAnalyticsRepository{},
			userRepo,
		)

		result, err := service.GetDashboardStats(ctx, userID, query)

		require.Nil(t, result)
		require.Error(t, err)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "USER_NOT_FOUND", appErr.Code)
	})

	t.Run("user has no expenses", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		expenseCountCalled := false

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpenseCountFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int, error) {
				expenseCountCalled = true
				return 0, nil
			},

			getHighestExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (*expense.Expense, error) {
				t.Fatal("GetHighestExpense should not be called")
				return nil, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetDashboardStats(ctx, userID, query)

		require.Nil(t, result)
		require.Error(t, err)
		require.True(t, expenseCountCalled)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "USER_HAS_NO_EXPENSES", appErr.Code)
	})

	t.Run("expense count repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getExpenseCountFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (int, error) {
				return 0, expectedErr
			},

			getHighestExpenseFn: func(ctx context.Context, gotUserID uuid.UUID, gotQuery *analytic.GetDashboardQuery) (*expense.Expense, error) {
				t.Fatal("GetHighestExpense should not be called")
				return nil, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetDashboardStats(ctx, userID, query)

		require.Nil(t, result)
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestExpenseAnalyticsService_GetSpendingTrends(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()

	expectedTrends := []analytic.MonthlyTotals{
		{
			Month:      "2026-06",
			TotalCents: 15000,
			Count:      4,
		},
		{
			Month:      "2026-07",
			TotalCents: 22000,
			Count:      6,
		},
	}

	t.Run("success", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				require.Equal(t, userID.String(), payload.ID)

				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getSpendingTrendsFn: func(ctx context.Context, gotUserID uuid.UUID) ([]analytic.MonthlyTotals, error) {
				require.Equal(t, userID, gotUserID)

				return expectedTrends, nil
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetSpendingTrends(ctx, userID)

		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, expectedTrends, result)
	})

	t.Run("user not found", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return nil, errors.New("user not found")
			},
		}

		service := NewExpenseAnalyticsService(
			nil,
			&fakeExpenseAnalyticsRepository{},
			userRepo,
		)

		result, err := service.GetSpendingTrends(ctx, userID)

		require.Nil(t, result)
		require.Error(t, err)

		var appErr *errs.AppError
		require.ErrorAs(t, err, &appErr)
		require.Equal(t, "USER_NOT_FOUND", appErr.Code)
	})

	t.Run("repository error", func(t *testing.T) {
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())

		expectedErr := errors.New("database error")

		userRepo := &fakeUserRepository{
			getUserByIDFn: func(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
				return &user.User{
					Base: model.Base{
						BaseWithID: model.BaseWithID{
							ID: userID,
						},
					},
				}, nil
			},
		}

		analyticsRepo := &fakeExpenseAnalyticsRepository{
			getSpendingTrendsFn: func(ctx context.Context, gotUserID uuid.UUID) ([]analytic.MonthlyTotals, error) {
				return nil, expectedErr
			},
		}

		service := NewExpenseAnalyticsService(nil, analyticsRepo, userRepo)

		result, err := service.GetSpendingTrends(ctx, userID)

		require.Nil(t, result)
		require.Error(t, err)
		require.ErrorIs(t, err, expectedErr)
	})
}
