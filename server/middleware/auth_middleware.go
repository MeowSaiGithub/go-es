package middlewares

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	logs "github.com/rs/zerolog/log"
	cErr "go-es/internal/errors"
	"go-es/internal/response"
	"net/http"
)

// Auth is a JWT authentication middleware
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the Authorization header exists
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logs.Debug().Msg("Authorization header not found")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusUnauthorized,
				Message: "Authorization token required",
				Details: "Authorization token required",
				Type:    cErr.UnauthorizedError.String(),
			})
			c.Abort()
			return
		}

		// Expect "Bearer <token>"
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			logs.Debug().Msg("Invalid authorization format")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid authorization format",
				Details: "Invalid authorization format",
				Type:    cErr.UnauthorizedError.String(),
			})
			c.Abort()
			return
		}

		// Extract the token part
		tokenStr := authHeader[7:]

		// Parse the token
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token's signing method is HMAC (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key to validate the JWT
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			logs.Debug().Err(err).Msg("Invalid token")
			response.SendErrorResponse(c, response.ErrResponse{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				Details: err.Error(),
				Type:    cErr.UnauthorizedError.String(),
			})
			c.Abort()
			return
		}

		// Token is valid, continue processing the request
		c.Next()
	}
}
