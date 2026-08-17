package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/repositories"
)

func registerUserRoutes(r gin.IRouter, h *handlers.UserHandlers, sessionRepo *repositories.SessionRepository, jwtSecret []byte) {
	// Register user
	users := r.Group("/users")
	users.GET("", h.GetUsers())
	users.GET("/:id", h.GetUserByID())

	protected := users.Group("")
	protected.Use(middleware.RequireAuth(sessionRepo, jwtSecret))
	protected.PATCH("/:id", h.UpdateUser())
	protected.DELETE("/:id", h.DeleteUser())
}
