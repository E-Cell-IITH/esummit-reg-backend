package server

import (
	"context"
	"net/http"
	constants "reg/internal/const"
	"reg/internal/cookies"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open routes that do not require authentication
		if c.Request.URL.Path == "/logout" || strings.HasPrefix(c.Request.URL.Path, "/signup") || strings.HasPrefix(c.Request.URL.Path, "/signin") || c.Request.URL.Path == "/health" || c.Request.URL.Path == "/register" || c.Request.URL.Path == "/update-startup-sheet" {
			c.Next()
			return
		}

		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing Authorization header"})
			c.Abort()
			return
		}

		// Check and strip the "Bearer" prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid token format"})
			c.Abort()
			return
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify token
		res, err := cookies.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid token"})
			c.Abort()
			return
		}

		// Add user information to the request context
		ctx := context.WithValue(c.Request.Context(), constants.UserIDKey, res.Subject)
		ctx = context.WithValue(ctx, constants.EmailKey, res.Email)
		c.Request = c.Request.WithContext(ctx)

		// Continue to the next middleware or handler
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, ok := c.Request.Context().Value(constants.UserIDKey).(string)
	return userID, ok
}
