package controllers

import (
	"fmt"
	"net/http"
	"reg/internal/database"
	email "reg/internal/emails"
	"reg/internal/utils"

	"github.com/gin-gonic/gin"
)

type OtpRequest struct {
	Email string `json:"email"`
}

func SendOtp(c *gin.Context) {
	var req OtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
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
