package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "X-Request-ID"

// RequestID is a middleware that assigns a unique request ID to each incoming request.
// It assigns the ID to the context as well as the response header.
// The ID is a RFC 4122-compliant UUID.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request already has a Request-ID
		requestID := c.GetHeader(RequestIDKey)
		if requestID == "" {
			// Generate a new UUID if none exists
			uid, _ := uuid.NewV7()
			requestID = uid.String()
		}

		// Add Request-ID to the context and response header
		c.Set(RequestIDKey, requestID)
		c.Writer.Header().Set(RequestIDKey, requestID)

		c.Next()
	}
}
