package controllers

import (
	"context"
	"net/http"
	"reg/internal/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

type contextKey string

const userIDKey contextKey = "userID"

func GetUserHandler(c *gin.Context) {
	// Get user id from context
	userid, ok := c.Request.Context().Value(userIDKey).(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing session cookie"})
		return
	}
	id, err := strconv.Atoi(userid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id"})
		return
	}

	user, err := database.GetUserById(context.Background(), int64(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, user)
}
