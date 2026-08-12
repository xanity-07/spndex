package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	"github.com/xanity-07/spndex/internal/server"
)

func NewRouter(s *server.Server, h *handlers.Handlers) *gin.Engine {
	mw := middleware.NewMiddlewares(s)

	router := gin.New()

	// Global middleware registration
	router.Use(
		mw.Global.CORS(),
		mw.Global.Secure(),
		middleware.RequestID(),
		mw.Tracing.NewRelicMiddleware(),
		mw.Tracing.EnhanceTracing(),
		mw.ContextEnhancer.EnhanceContext(),
		mw.Global.RequestLogger(),
		mw.Global.Recover(),
	)

	// register system routes
	registerSystemRoutes(router, h)

	// register versioned routes
	router.Group("api/v1")

	return router
}
