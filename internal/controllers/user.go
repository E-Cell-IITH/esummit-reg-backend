package controllers

import (
	"context"
	"fmt"
	"net/http"
	constants "reg/internal/const"
	"reg/internal/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserHandler(c *gin.Context) {
	// Get user id from context
	fmt.Println(c.Request.Context())
	userid, ok := c.Request.Context().Value(constants.UserIDKey).(string)
	if !ok {
		fmt.Println("User ID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing session cookie"})
		return
	}
	id, err := strconv.Atoi(userid)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id"})
		return
	}

	user, ticketId, err := database.GetMeUser(context.Background(), int64(id))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User found",
		"user":     user,
		"ticketId": ticketId,
	})

}

func LogoutHandler(c *gin.Context) {

	// c.SetCookie("session", "", 1, "/", "", false, true)
	// cookies.SetCookie(c.Writer, "session", "l", 0)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}
