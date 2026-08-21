package repositories

import "github.com/xanity-07/spndex/internal/server"

type Repositories struct {
	ExpenseRepo          *ExpenseRepository
	UserRepo             *UserRepository
	SessionRepo          *SessionRepository
	ExpenseAnalyticsRepo *ExpenseAnalyticsRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		ExpenseRepo:          NewExpenseRepository(s),
		UserRepo:             NewUserRepository(s),
		SessionRepo:          NewSessionRepository(s),
		ExpenseAnalyticsRepo: NewAnalyticsRepository(s),
	}
}
