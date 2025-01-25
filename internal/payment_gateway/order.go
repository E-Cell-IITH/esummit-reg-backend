package paymentgateway

import (
	"context"
	"fmt"
	"net/http"
	constants "reg/internal/const"
	"reg/internal/database"
	emails "reg/internal/emails"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaymentInitiate struct {
	Amount float64 `json:"amount"`
	TxnId  string  `json:"txn_id"`
	Title  string  `json:"title"`
	IsAccommodation bool `json:"isAccommodation"`
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

	user, err := database.GetUserById(context.Background(), int64(userIdInt))
	if err != nil {
		fmt.Println("TAKE ACTION>>>>>>>>>>>>>>>>>>> FOR ID: ", userIdInt)
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}

	// if amount is -1
	if req.Amount == -1 {
		err := database.AddBasicTickets(userIdInt, req.Title)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tickets", "err": err})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Tickets purchased successfully"})

		//SEND EMAIL
		data, err := emails.LoadPurchasedTicketTemplate(user.Name, "STANDARD", "Free")
		if err != nil {
			fmt.Println(err)
			fmt.Println("TAKE ACTION>>>>>>>>>>>>>>>>>>> FOR ID: ", userIdInt)
		}

		emails.SendEmail(user.Email, nil, "Your E-Summit 2025 Pass Confirmation", data, "")
		return
	}

	id, err := database.CreatePaymentRecord(req.TxnId, userIdInt, req.Amount, req.Title, req.IsAccommodation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push transaction ID"})
		return
	}

	if id == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction ID already exists"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction ID add successfully", "payment_id": id})
	//SEND EMAIL
	data, err := emails.LoadPendingTemplate(user.Name, req.TxnId, fmt.Sprintf("%.2f", req.Amount))
	if err != nil {
		fmt.Println(err)
		fmt.Println("TAKE ACTION>>>>>>>>>>>>>>>>>>> FOR ID: ", userIdInt)
	}

	emails.SendEmail(user.Email, nil, "Payment Confirmation Pending for E-Summit 2025", data, "")
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

	if res == 45 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already verified"})
		return
	}

	//Update tickets table
	title, err := database.AddTickets(id, req.TxnId)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tickets", "err": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction ID verified successfully", "userId": id})
	//SEND EMAIL
	user, err := database.GetUserById(context.Background(), int64(id))

	if err != nil {
		fmt.Println(err)
		fmt.Println("TAKE ACTION>>>>>>>>>>>>>>>>>>> FOR ID: ", id)
	}

	data, err := emails.LoadPurchasedTicketTemplate(user.Name, title, fmt.Sprintf("%.2f", req.Amount))
	if err != nil {
		fmt.Println(err)
		fmt.Println("TAKE ACTION>>>>>>>>>>>>>>>>>>> FOR ID: ", id)
	}

	emails.SendEmail(user.Email, nil, "Your E-Summit 2025 Pass Confirmation", data, "")
}
