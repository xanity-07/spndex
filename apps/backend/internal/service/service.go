package service

import (
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
)

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
