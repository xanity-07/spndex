package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
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
		enums.BindingJSON,
	)
}

func (h *UserHandlers) GetUsers() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, query *user.GetUsersQuery) (*model.PaginatedResponse[user.User], error) {
		return h.userService.GetUsers(c, query)
	},
		http.StatusOK,
		&user.GetUsersQuery{},
		enums.BindingQuery,
	)
}

func (h *UserHandlers) GetUserByID() gin.HandlerFunc {
	return Handle(h.Handler,
		func(c *gin.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
			return h.userService.GetUserByID(c, payload)
		}, http.StatusOK,
		&user.GetUserByIDPayload{},
		enums.BindingURI,
	)
}

func (h *UserHandlers) UpdateUser() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *user.UpdateUserPayload) (*user.User, error) {
			id := c.Param("id")
			userID := uuid.MustParse(id)
			return h.userService.UpdateUser(c, userID, payload)
		},
		http.StatusOK,
		&user.UpdateUserPayload{},
		enums.BindingJSON,
	)
}

func (h *UserHandlers) DeleteUser() gin.HandlerFunc {
	return HandleNoContent(h.Handler, func(c *gin.Context, payload *user.DeleteUserPayload) error {
		return h.userService.DeleteUser(c, payload)
	},
		http.StatusNoContent,
		&user.DeleteUserPayload{},
		enums.BindingURI,
	)
}
