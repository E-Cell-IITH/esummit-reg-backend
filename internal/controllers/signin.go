package controllers

import (
	"context"
	"fmt"
	"net/http"
	"reg/internal/cookies"
	"reg/internal/database"
	email "reg/internal/emails"
	"reg/internal/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

func SendOtpSignIN(c *gin.Context) {
	var req OtpRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	req.Email = strings.ToLower(req.Email)

	// if user not exists
	if !database.UserExists(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exists"})
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

	_, err = email.SendEmail(req.Email, nil, "OTP Verification for E-Summit-2025 | E-Cell IIT Hyderabad", body, "")
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

func VerifyOtpSignIN(c *gin.Context) {
	var req OtpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Otp == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	req.Email = strings.ToLower(req.Email)

	// Verify User
	if !database.UserExists(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User does not exists"})
		return
	}

	// 1. Verify OTP
	if !database.VerifyOtp(req.Email, req.Otp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid OTP"})
		return
	}

	// 2. Get user ID
	user, err := database.GetUserByEmail(context.Background(), req.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// 3. Generate JWT
	token, err := cookies.GenerateToken(user.ID, req.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// cookies.SetCookie(c.Writer, "session", token, 0)

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully", "token": token})
}
