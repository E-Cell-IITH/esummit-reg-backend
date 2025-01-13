package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"reg/internal/database"
	email "reg/internal/emails"
	"reg/internal/model"

	firebase "firebase.google.com/go"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

func RegisterHandler(c *gin.Context) {
	// 1. Parse incoming JSON
	var req model.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Verify Firebase ID Token
	fmt.Println(req.Token)

	// Initialize Firebase app
	app, err := firebase.NewApp(context.Background(), nil, option.WithCredentialsFile("serviceAccountKey.json"))
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Firebase"})
		return
	}

	// Get Auth client from Firebase App
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Firebase Auth client: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize Firebase Auth client"})
		return
	}

	// Verify the ID token using Firebase Auth
	token, err := client.VerifyIDToken(context.Background(), req.Token)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase ID token"})
		return
	}

	googleUserID := token.UID
	fmt.Printf("Firebase User ID: %s\n", googleUserID)

	// 3. Save registration data
	id, err := database.CreateRegistration(context.Background(), req.Data)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	

	// Send email after this
	// 5. Send email
	body, err := email.LoadRegistrationTemplate(req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}
	_, err = email.SendEmail(req.Data.Email, nil, "Registration Successful for Startup Fair 2025 | E-Cell IIT Hyderabad", body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// 4. Send success response
	c.JSON(http.StatusOK, gin.H{
		"message":      "Registration successful",
		"firebaseUser": googleUserID,
		"id":           id,
	})
}
