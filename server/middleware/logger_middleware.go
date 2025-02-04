package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"time"
)

// Logger returns a gin middleware that logs the request details
//
// It uses the X-Request-ID header value to associate the log entries with the request.
// If no X-Request-ID is provided, it will use the "unknown" string as the request ID.
//
// It logs the following information:
//
//   - The request timestamp
//   - The request method
//   - The request path
//   - The request ID
//   - The response status code
//   - The time taken to process the request
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Retrieve the Request-ID from the context
		requestID := c.GetString(RequestIDKey)
		// This should not happen at all
		if requestID == "" {
			requestID = "unknown"
		}

		logger := log.With().
			Timestamp().
			Str("request_id", requestID).
			Str("method", method).
			Str("path", path).
			Logger()

		c.Set("logger", logger)

		// Process the request
		c.Next()

		// After request processing
		duration := time.Since(start)
		status := c.Writer.Status()

		// Log the request details with Request-ID
		log.Info().
			Int("status", status).
			Dur("duration", duration).
			Msg("request processed")
	}
}
