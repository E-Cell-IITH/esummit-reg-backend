package controllers

import (
	"context"
	"fmt"
	"net/http"
	"reg/internal/database"
	"reg/internal/model"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/idtoken"
)

func RegisterHandler(c *gin.Context) {
	// 1. Parse incoming JSON
	var req model.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Verify Google ID Token
	payload, err := idtoken.Validate(context.Background(), req.Token, "GOOGLE_CLIENT_ID")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google ID token"})
		return
	}

	googleUserID := payload.Subject
	fmt.Printf("Google User ID: %s\n", googleUserID)

	// 3. Save registration data
	id, err := database.CreateRegistration(context.Background(), req.Data)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Sever Error"})
		return
	}

	// 4. Send success response
	c.JSON(http.StatusOK, gin.H{
		"message":    "Registration successful",
		"googleUser": googleUserID,
		"id":         id,
	})
}
