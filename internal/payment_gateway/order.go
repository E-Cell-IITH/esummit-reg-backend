package paymentgateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"reg/internal/database"
	"reg/internal/server"
	"strconv"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

type PaymentInitiate struct {
	Amount float64 `json:"amount"`
}

func CreateOrder(c *gin.Context) {
	var req PaymentInitiate
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	userId, ok := server.GetUserID(c)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error! While fetching user"})
		return
	}

	client := razorpay.NewClient(os.Getenv("RAZORPAY_API_KEY"), os.Getenv("RAZORPAY_API_SECRET"))
	cnt, err := database.GetCountOfOrders()
	if err != nil {
		fmt.Println("error in getting count of orders", err)
		cnt = 0 // set some random number
	}
	cnt = cnt + 1
	receipt := fmt.Sprintf("order#%d", cnt)
	data := map[string]interface{}{
		"amount":   req.Amount,
		"currency": "INR",
		"receipt":  receipt,
	}

	order, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error! While initiating payment"})
		return
	}

	orderID, ok := order["id"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error! Invalid order ID"})
		return
	}
	_, err = database.CreateOrder(int64(userIdInt), orderID, req.Amount, receipt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error! While saving order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "receipt": receipt, "contact_number": user.ContactNumber, "name": user.Name, "email": user.Email})
}
