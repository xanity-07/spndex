package repositories

import (
	"github.com/xanity-07/spndex/internal/server"
)

type ExpenseRepository struct {
	server *server.Server
}

func NewExpenseRepository(s *server.Server) *ExpenseRepository {
	return &ExpenseRepository{
		server: s,
	}
}
