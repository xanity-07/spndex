package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xanity-07/spndex/internal/server"
)

type OpenAPIHandler struct {
	Handler
}

func NewOpenAPIHandler(s *server.Server) *OpenAPIHandler {
	return &OpenAPIHandler{
		Handler: NewHandler(s),
	}
}

func (h *OpenAPIHandler) ServeOpenAPIUI(c *gin.Context) {
	templateBytes, err := os.ReadFile("static/openapi.html")
	if err != nil {
		h.server.Logger.Error().Msgf("failed to serve OpenAPI UI template: %s", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// do not cache the file so we can see new changes with refresh
	c.Writer.Header().Set("Cache-Control", "no-store")
	c.Data(http.StatusOK, "text/html; charset=utf-8", templateBytes)
}
