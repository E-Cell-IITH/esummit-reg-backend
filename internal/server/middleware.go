package server

import (
	"context"
	"net/http"
	"reg/internal/cookies"
	"strings"

	"github.com/gin-gonic/gin"
)

type contextKey string

const (
	userIDKey contextKey = "userID"
	emailKey  contextKey = "email"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open routes that do not require authentication
		if strings.HasPrefix(c.Request.URL.Path, "/signup") || strings.HasPrefix(c.Request.URL.Path, "/signin") || c.Request.URL.Path == "/health" || c.Request.URL.Path == "/register" || c.Request.URL.Path == "/update-startup-sheet" || c.Request.URL.Path == "/" {
			c.Next()
			return
		}

		// Get session cookie
		cookie, err := c.Cookie("session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing session cookie"})
			c.Abort()
			return
		}

		// Verify session cookie
		res, err := cookies.ParseToken(cookie)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid session"})
			c.Abort()
			return
		}

		// Add user information to the request context
		ctx := context.WithValue(c.Request.Context(), userIDKey, res.Subject)
		ctx = context.WithValue(ctx, emailKey, res.Email)
		c.Request = c.Request.WithContext(ctx)

		// Continue to the next middleware or handler
		c.Next()
	}
}

func GetUserID(c *gin.Context) (string, bool) {
	userID, ok := c.Request.Context().Value(userIDKey).(string)
	return userID, ok
}