package paymentgateway

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleCopounVerifications(c *gin.Context) {
	code := c.Param("code")
	coupons := os.Getenv("COUPON_CODES")
	couponMap := make(map[string]int)

	for _, coupon := range strings.Split(coupons, ",") {
		parts := strings.Split(strings.TrimSpace(coupon), ":")
		if len(parts) == 2 {
			discount, err := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err == nil {
				couponMap[strings.TrimSpace(parts[0])] = discount
			}
		}
	}

	if discount, exists := couponMap[code]; exists {
		c.JSON(http.StatusOK, gin.H{"code": code, "discount": discount})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Coupon code"})
	}
}
