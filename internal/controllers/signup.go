package controllers

import (
	"context"
	"fmt"
	"net/http"
	"reg/internal/database"
	email "reg/internal/emails"
	"reg/internal/model"
	"reg/internal/utils"

	"github.com/gin-gonic/gin"
)

type OtpRequest struct {
	Email string `json:"email"`
}
type OtpVerifyRequest struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

func SendOtpSignUP(c *gin.Context) {
	var req OtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// if user already exists
	if database.UserExists(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}

	// 1. Generate OTP
	otp := utils.GenerateOtp()

	// 2. Save OTP in database
	err := database.SaveOtp(req.Email, otp)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// 3. Send OTP via email
	body, err := email.LoadOtpVerificationsTemplate(otp)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	_, err = email.SendEmail(req.Email, nil, "OTP Verification for E-Summit-2025 | E-Cell IIT Hyderabad", body)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

func VerifyOtpSignUP(c *gin.Context) {
	var req OtpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 1. Verify OTP
	if !database.VerifyOtp(req.Email, req.Otp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

func RegisterUserHandler(c *gin.Context) {
	// 1. Parse incoming JSON
	var req model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Check if the user already exists
	if database.UserExists(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return

	}

	// 3. Save user data
	id, err := database.CreateUser(context.Background(), req)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"id":      id,
	})
}