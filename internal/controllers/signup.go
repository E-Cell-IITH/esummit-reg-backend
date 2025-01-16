package controllers

import (
	"context"
	"fmt"
	"net/http"
	"reg/internal/cookies"
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
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
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
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Otp == "" {
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
	type User struct {
		Email string `json:"email"`
		Name  string `json:"name"`
		Data  string `json:"data"`
		Otp   string `json:"otp"`
	}

	var req User
	if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" || req.Name == "" || req.Otp == "" {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Check if the user already exists
	if database.UserExists(req.Email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return

	}

	// 3. Check weather the OTP is verified
	if !database.VerifyOtp(req.Email, req.Otp) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP not verified"})
		return
	}

	// 4. Save user data
	id, err := database.CreateUser(context.Background(), model.User{
		Email: req.Email,
		Name:  req.Name,
		Data:  req.Data,
	})
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// 5. Generate Token
	token, err := cookies.GenerateToken(int(id), req.Email)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// 5. Set cookie
	cookies.SetCookie(c.Writer, "session", token)

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"id":      id,
	})

	// 6. Update the OTP status
	database.UpdateOtpStatus(req.Email)

	// 7. Send welcome email
	// TODO: Implement this
}
