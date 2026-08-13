package repositories

import "github.com/xanity-07/spndex/internal/server"

type Repositories struct {
	ExpenseRepo *ExpenseRepository
	UserRepo    *UserRepository
}

func NewRepositories(s *server.Server) *Repositories {
	return &Repositories{
		ExpenseRepo: NewExpenseRepository(s),
		UserRepo:    NewUserRepository(s),
	}
}
