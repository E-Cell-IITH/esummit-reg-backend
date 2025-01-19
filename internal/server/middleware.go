package server

import (
	"context"
	"fmt"
	"net/http"
	constants "reg/internal/const"
	"reg/internal/cookies"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Open routes that do not require authentication
		if strings.HasPrefix(c.Request.URL.Path, "/signup") || strings.HasPrefix(c.Request.URL.Path, "/signin") || c.Request.URL.Path == "/health" || c.Request.URL.Path == "/register" || c.Request.URL.Path == "/update-startup-sheet" {
			c.Next()
			return
		}

		// Get session cookie
		cookie, err := c.Cookie("session")
		fmt.Println(cookie)
		if err != nil {
			fmt.Println(err)
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

		fmt.Println(res.Subject)
		fmt.Println(res.Email)

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
