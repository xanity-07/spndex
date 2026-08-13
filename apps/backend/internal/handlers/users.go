package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/model"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type UserHandlers struct {
	Handler
	userService *service.UserService
}

func NewUserHandlers(s *server.Server, userSerivice *service.UserService) *UserHandlers {
	return &UserHandlers{
		Handler:     NewHandler(s),
		userService: userSerivice,
	}
}

func (h *UserHandlers) CreateUser() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *user.CreateUserPayload) (*user.User, error) {
			return h.userService.CreateUser(c, payload)
		},
		http.StatusCreated,
		&user.CreateUserPayload{},
	)
}

func (h *UserHandlers) GetUsers() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
		return h.userService.GetUsers(c, query)
	},
		http.StatusOK,
		&user.GetUsersQuery{},
	)
}
