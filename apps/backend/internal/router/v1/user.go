package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
)

func registerUserRoutes(r gin.IRouter, h *handlers.UserHandlers) {
	// Register user
	users := r.Group("/users")
	users.GET("", h.GetUsers())
	users.GET("/:id", h.GetUserByID())
	users.PATCH("/:id", h.UpdateUser())
	users.DELETE("/:id", h.DeleteUser())
}
