package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/repositories"
)

func registerAuthRoutes(r gin.IRouter, h *handlers.AuthHandler, sessionRepo *repositories.SessionRepository, jwtSecret []byte) {
	auth := r.Group("/auth")
	auth.POST("/register", h.Register())
	auth.POST("/login", h.Login())

	protected := auth.Group("")

	protected.Use(middleware.RequireAuth(sessionRepo, jwtSecret))
	protected.POST("logout", h.Logout())
}
