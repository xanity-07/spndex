package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"
)

// RequestID connects the whole request lifecycle to connect all the different interactions
// into one transaction
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Set the RequestID in our context
		c.Set(RequestIDKey, requestID)
		// Set X-Request-ID header to the newly generated RequestID
		c.Writer.Header().Set(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID is a helper function that returns the request ID from our context if it exists
func GetRequestID(c *gin.Context) string {
	if requestID, ok := c.Get(RequestIDKey); ok {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
