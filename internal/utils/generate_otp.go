package utils

import (
	"fmt"
	"time"

	"golang.org/x/exp/rand"
)

func GenerateOtp() string {
	// Generate a random 6-digit OTP
	rand.Seed(uint64(time.Now().UnixNano()))
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	return otp
}
