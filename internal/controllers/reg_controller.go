package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"reg/internal/config"
	"reg/internal/database"
	email "reg/internal/emails"
	"reg/internal/model"
	"strings"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func RegisterHandler(c *gin.Context) {
	// 1. Parse incoming JSON
	var req model.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Token == "" || req.Data.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// 2. Verify Firebase ID Token
	fmt.Println(req.Token)

	// Verify the ID token using Firebase Auth
	token, err := config.Client.VerifyIDToken(context.Background(), req.Token)
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
	_, err = email.SendEmail(req.Data.Email, nil, "Registration Successful for Startup Fair 2025 | E-Cell IIT Hyderabad", body, os.Getenv("SMTP_REPLY_TO"))
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

func PostDataInGSheet(c *gin.Context) {
	type Data struct {
		IdToken string `json:"token"`
	}
	var req Data
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	token, err := config.Client.VerifyIDToken(context.Background(), req.IdToken)

	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase ID token"})
		return
	}
	email := token.Claims["email"].(string)
	admins := os.Getenv("ADMIN_EMAILS")

	// Check if the email is an admin
	if !isAdmin(email, admins) {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to access this resource"})
		return
	}

	reg_data, err := database.GetRegistrationsYetToPush(context.Background())
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	if err := writeToGSheet(reg_data); err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write data to Google Sheets"})
		return
	}

	// Mark the registrations as pushed
	err = database.MarkRegistrationAsPushed(context.Background(), reg_data)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark registrations as pushed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Data written to Google Sheets successfully"})

}

func isAdmin(email, admins string) bool {
	adminList := strings.Split(admins, ",")
	for _, adminEmail := range adminList {
		if adminEmail == email {
			return true
		}
	}
	return false
}

func writeToGSheet(data []model.RegistrationData) error {
	clientOption := option.WithCredentialsFile("service-account.json")
	srv, err := sheets.NewService(context.Background(), clientOption)
	if err != nil {
		return fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	spreadsheetId := os.Getenv("GOOGLE_SHEET_ID")
	writeRange := "Sheet1!A1"

	var vr sheets.ValueRange
	for _, reg := range data {
		var row []interface{}
		val := reflect.ValueOf(reg)
		for i := 0; i < val.NumField(); i++ {
			row = append(row, val.Field(i).Interface())
		}
		vr.Values = append(vr.Values, row)
	}

	_, err = srv.Spreadsheets.Values.Append(spreadsheetId, writeRange, &vr).
		ValueInputOption("RAW").
		InsertDataOption("INSERT_ROWS").
		Do()

	if err != nil {
		return fmt.Errorf("unable to write data to sheet: %v", err)
	}

	return nil
}
