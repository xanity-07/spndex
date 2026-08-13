package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
)

func registerUserRoutes(r gin.IRouter, h *handlers.UserHandlers) {
	// Register user
	users := r.Group("/users")
	users.POST("", h.CreateUser())
	users.GET("", h.GetUsers())
}
