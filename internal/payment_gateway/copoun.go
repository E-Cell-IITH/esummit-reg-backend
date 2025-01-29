package paymentgateway

import (
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandleCouponVerifications(c *gin.Context) {
	coupons := os.Getenv("COUPON_CODES")

	type CouponDetails struct {
		Discount      int     // Discount amount
		OriginalPrice float64 // Original price the coupon applies to
	}

	couponMap := make(map[string]CouponDetails)

	var requestBody struct {
		Code          string  `json:"couponCode"`
		OriginalPrice float64 `json:"originalPrice"`
	}

	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	code := strings.TrimSpace(requestBody.Code)
	requestOriginalPrice := requestBody.OriginalPrice

	for _, coupon := range strings.Split(coupons, ",") {
		coupon = strings.TrimSpace(coupon)
		parts := strings.Split(coupon, ":")
		if len(parts) == 2 {
			couponCode := strings.TrimSpace(parts[0])
			discountAndPrice := strings.TrimSpace(parts[1])
			dpParts := strings.Split(discountAndPrice, ";")
			if len(dpParts) == 2 {
				discountStr := strings.TrimSpace(dpParts[0])
				originalPriceStr := strings.TrimSpace(dpParts[1])

				// Convert discount and original price strings to appropriate types
				discount, err1 := strconv.Atoi(discountStr)
				originalPrice, err2 := strconv.ParseFloat(originalPriceStr, 64)
				if err1 == nil && err2 == nil {
					// Store the coupon details in the map
					couponMap[couponCode] = CouponDetails{
						Discount:      discount,
						OriginalPrice: originalPrice,
					}
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon"})
					return
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon"})
				return
			}
		}
	}

	if couponInfo, exists := couponMap[code]; exists {
		if couponInfo.OriginalPrice == requestOriginalPrice {
			newPrice := requestOriginalPrice - float64(couponInfo.Discount)
			c.JSON(http.StatusOK, gin.H{
				"code":          code,
				"discount":      couponInfo.Discount,
				"newPrice":      newPrice,
				"originalPrice": requestOriginalPrice,
				"message":       "Congratulations! You Saved " + strconv.Itoa(couponInfo.Discount) + " on this purchase",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon does not apply to this pass"})
		}
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid Coupon code"})
	}
}