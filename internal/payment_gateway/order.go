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

func CreateOrder(c *gin.Context) {
	var req PaymentInitiate
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount == 0 {
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

	c.JSON(http.StatusOK, gin.H{"message": "Transaction ID add successfully", "payment_id": id})
}
