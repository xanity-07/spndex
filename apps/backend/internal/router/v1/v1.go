package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
)

func RegisterV1Routes(router gin.IRouter, handlers *handlers.Handlers, middleware *middleware.Middlewares) {
	// Register user routes
	registerUserRoutes(router, handlers.User)
}
