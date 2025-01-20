package paymentgateway

import (
	"context"
	"fmt"
	"net/http"
	constants "reg/internal/const"
	"reg/internal/database"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentInitiate struct {
	Amount float64 `json:"amount"`
	TxnId  string  `json:"txn_id"`
}

func getUserID(c *gin.Context) (string, bool) {
	userID, ok := c.Request.Context().Value(constants.UserIDKey).(string)
	return userID, ok
}
func getEmail(c *gin.Context) (string, bool) {
	email, ok := c.Request.Context().Value(constants.EmailKey).(string)
	return email, ok
}

func CreateOrder(c *gin.Context) {
	var req PaymentInitiate
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount == 0 {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	userId, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing user ID"})
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid user ID"})
		return
	}

	user, ticketId, err := database.GetMeUser(context.Background(), int64(userIdInt))
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		return
	}

	// Create a new order
	id, err := database.InitiatePayment(req.Amount, userIdInt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order_id": id, "message": "User found",
		"user":     user,
		"ticketId": ticketId})
}

func PushTransactionIds(c *gin.Context) {
	var req PaymentInitiate
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount == 0 || req.TxnId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	userId, ok := getUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing user ID"})
		return
	}

	userIdInt, err := strconv.Atoi(userId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid user ID"})
		return
	}

	id, err := database.CreatePaymentRecord(req.TxnId, userIdInt, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push transaction ID"})
		return
	}

	if id == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Transaction ID already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction ID add successfully", "payment_id": id})
}

func AddSuccessfulTxnIds(c *gin.Context) {
	email, ok := getEmail(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing user email"})
		return
	}

	if email != "ADMIN" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: invalid email"})
		return
	}

	var req PaymentInitiate
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount == 0 || req.TxnId == "" {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	res, id, err := database.AddSuccessfulTxnIds(req.TxnId, req.Amount)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add transaction ID"})
		return
	}
	if res == -1 {
		c.JSON(http.StatusAccepted, gin.H{"message": "No transaction ID found in the database", "userId": id})
		return
	}

	if res ==45 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already verified"})
		return
	}

	//Update tickets table
	err = database.AddTickets(id, req.Amount)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction ID verified successfully", "userId": id})
}