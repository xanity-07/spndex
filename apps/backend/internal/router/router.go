package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/handlers"
	"github.com/xanity-07/spndex/internal/middleware"
	v1 "github.com/xanity-07/spndex/internal/router/v1"
	"github.com/xanity-07/spndex/internal/server"
)

func NewRouter(s *server.Server, h *handlers.Handlers) *gin.Engine {
	mw := middleware.NewMiddlewares(s)

	router := gin.New()

	// Global middleware registration
	router.Use(
		mw.Global.Recover(),
		mw.Global.CORS(),
		mw.Global.Secure(),
		middleware.RequestID(),
		mw.Tracing.NewRelicMiddleware(),
		mw.Tracing.EnhanceTracing(),
		mw.ContextEnhancer.EnhanceContext(),
		mw.Global.RequestLogger(),
	)
	// register system routes
	registerSystemRoutes(router, h)

	// register versioned routes
	v1Router := router.Group("/api/v1")
	v1.RegisterV1Routes(v1Router, h, mw)

	return router
}
