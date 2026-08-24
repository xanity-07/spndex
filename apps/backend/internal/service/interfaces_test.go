package service

import (
	"testing"

	"github.com/xanity-07/spndex/internal/repositories"
)

func TestRepositoryImplementations(t *testing.T) {
	var _ ExpenseRepository = (*repositories.ExpenseRepository)(nil)
	var _ UserRepository = (*repositories.UserRepository)(nil)
	var _ SessionRepository = (*repositories.SessionRepository)(nil)
}
