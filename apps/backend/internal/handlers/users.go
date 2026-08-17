package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/errs"
	"github.com/xanity-07/spndex/internal/middleware"
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

func (h *UserHandlers) GetUsers() gin.HandlerFunc {
	return Handle(h.Handler,
		func(
			c *gin.Context,
			query *user.GetUsersQuery,
		) (*model.PaginatedResponse[user.User], error) {
			return h.userService.GetUsers(c, query)
		},
		http.StatusOK,
		func() *user.GetUsersQuery {
			return &user.GetUsersQuery{}
		},
		enums.BindingQuery,
	)
}

func (h *UserHandlers) GetUserByID() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *user.GetUserByIDPayload) (*user.User, error) {
			return h.userService.GetUserByID(c, payload)
		},
		http.StatusOK,
		func() *user.GetUserByIDPayload {
			return &user.GetUserByIDPayload{}
		},
		enums.BindingURI,
	)
}

func (h *UserHandlers) UpdateUser() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *user.UpdateUserPayload) (*user.User, error) {
			id := c.Param("id")
			userID := uuid.MustParse(id)

			val, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			authID, ok := val.(uuid.UUID)
			if !ok {
				return nil, errs.NewUnauthorizedError("unauthorized", false)
			}

			if authID != userID {
				return nil, errs.NewUnauthorizedError("invalid operation", false)
			}

			return h.userService.UpdateUser(c, userID, payload)
		},
		http.StatusOK,
		func() *user.UpdateUserPayload {
			return &user.UpdateUserPayload{}
		},
		enums.BindingJSON,
	)
}

func (h *UserHandlers) DeleteUser() gin.HandlerFunc {
	return HandleNoContent(
		h.Handler,
		func(c *gin.Context, payload *user.DeleteUserPayload) error {

			val, ok := c.Get(middleware.UserIDKey)
			if !ok {
				return errs.NewUnauthorizedError("unauthorized 1", false)
			}

			authID, ok := val.(uuid.UUID)
			if !ok {
				return errs.NewUnauthorizedError("unauthorized 2", false)
			}

			if authID.String() != payload.ID {
				return errs.NewUnauthorizedError("invalid operation", false)
			}

			return h.userService.DeleteUser(c, payload)
		},
		http.StatusNoContent,
		&user.DeleteUserPayload{},
		enums.BindingURI,
	)
}
