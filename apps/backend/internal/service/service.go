package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/analytic.go"
	"github.com/xanity-07/spndex/internal/model/expense"
	"github.com/xanity-07/spndex/internal/model/session"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseRepository interface {
	CreateExpense(ctx context.Context, userID uuid.UUID, payload *expense.CreateExpensePayload) (*expense.Expense, error)
	DeleteExpense(ctx context.Context, userID uuid.UUID, payload *expense.DeleteExpense) error
	GetExpenseByID(ctx context.Context, userID uuid.UUID, payload *expense.GetExpenseByID) (*expense.Expense, error)
	GetExpenses(ctx context.Context, userID uuid.UUID, query *expense.GetExpensesQuery) (*model.PaginatedResponse[expense.Expense], error)
	UpdateExpense(ctx context.Context, userID uuid.UUID, expenseID uuid.UUID, payload *expense.UpdateExpense) (*expense.Expense, error)
}

type UserRepository interface {
	CheckUserExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, payload *user.CreateUserPayload) (*user.User, error)
	DeleteUser(ctx context.Context, payload *user.DeleteUserPayload) error
	GetUserByEmail(ctx context.Context, email string) (*user.User, error)
	GetUserByID(ctx context.Context, payload *user.GetUserByIDPayload) (*user.User, error)
	GetUsers(ctx context.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error)
	UpdateUser(ctx context.Context, userID uuid.UUID, payload *user.UpdateUserPayload) (*user.User, error)
}

type ExpenseAnalyticsRepository interface {
	GetAverageExpenseAmount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error)
	GetExpenseCount(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int, error)
	GetExpensesByCategory(ctx context.Context, userID uuid.UUID) ([]analytic.CategoryTotals, error)
	GetHighestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error)
	GetLowestExpense(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (*expense.Expense, error)
	GetMonthlyExpenses(ctx context.Context, userID uuid.UUID, payload *analytic.GetMonthlyExpensesPayload) ([]analytic.MonthlyTotals, error)
	GetSpendingTrends(ctx context.Context, userID uuid.UUID) ([]analytic.MonthlyTotals, error)
	GetTotalExpenses(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error)
	MonthlyTotals(ctx context.Context, userID uuid.UUID, query *analytic.GetDashboardQuery) (int64, error)
}

type SessionRepository interface {
	Create(ctx context.Context, session *session.Session, ttl time.Duration) error
	Delete(ctx context.Context, sessionID string) error
	Get(ctx context.Context, sessionID string) (*session.Session, error)
}

type Services struct {
	User             *UserService
	Auth             *AuthService
	Expense          *ExpenseService
	ExpenseAnalytics *ExpenseAnalyticsService
}

func NewServices(s *server.Server, repos *repositories.Repositories) *Services {
	return &Services{
		Auth:             NewAuthService(s, repos.UserRepo, repos.SessionRepo),
		User:             NewUserService(s, repos.UserRepo),
		Expense:          NewExpenseService(s, repos.ExpenseRepo, repos.UserRepo),
		ExpenseAnalytics: NewExpenseAnalyticsService(s, repos.ExpenseAnalyticsRepo, repos.UserRepo),
	}
}
