package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/enums"
	"github.com/xanity-07/spndex/internal/model/authmodel"
	"github.com/xanity-07/spndex/internal/model/user"
	"github.com/xanity-07/spndex/internal/server"
	"github.com/xanity-07/spndex/internal/service"
)

type AuthHandler struct {
	Handler
	AuthService *service.AuthService
}

func NewAuthHandler(s *server.Server, authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		Handler:     NewHandler(s),
		AuthService: authService,
	}
}

func (h *AuthHandler) Register() gin.HandlerFunc {
	return Handle(
		h.Handler,
		func(c *gin.Context, payload *user.CreateUserPayload) (*user.User, error) {
			return h.AuthService.Register(c, payload)
		},
		http.StatusCreated,
		func() *user.CreateUserPayload {
			return &user.CreateUserPayload{}
		},
		enums.BindingJSON,
	)
}

func (h *AuthHandler) Login() gin.HandlerFunc {
	return Handle(h.Handler, func(c *gin.Context, payload *authmodel.LoginPayload) (*authmodel.LoginResponsePayload, error) {
		return h.AuthService.Login(c, payload)
	},
		http.StatusOK,
		func() *authmodel.LoginPayload {
			return &authmodel.LoginPayload{}
		},
		enums.BindingJSON,
	)
}

func (h *AuthHandler) Logout() gin.HandlerFunc {
	return HandleNoContent(h.Handler,
		func(c *gin.Context, noReq EmptyRequest) error {
			return h.AuthService.Logout(c)
		}, http.StatusNoContent, EmptyRequest{}, enums.BindingJSON)
}
