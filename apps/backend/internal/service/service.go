package service

import (
	"github.com/xanity-07/spndex/internal/repositories"
	"github.com/xanity-07/spndex/internal/server"
)

type Services struct {
	User *UserService
}

func NewServices(s *server.Server, repos *repositories.Repositories) *Services {
	return &Services{
		User: NewUserService(s, repos.UserRepo),
	}
}
