package database

import (
	"fmt"
	"log"
)

// SaveOtp saves or updates an OTP in the database for a given email.
func SaveOtp(email string, otp string) error {
	// Save OTP in the "otps" table
	_, err := db.Exec(`
    INSERT INTO otps (email, otp, updated_at, is_expired) 
    VALUES (?, ?, DATETIME('now', 'localtime'), FALSE)
    ON CONFLICT(email) DO UPDATE SET 
        otp = excluded.otp, 
		is_expired = FALSE,
        updated_at = DATETIME('now', 'localtime')
	`, email, otp)


	if err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}

	return nil
}

func VerifyOtp(email, otp string) bool {
	// Check if the OTP is valid
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM otps 
			WHERE email = ? AND otp = ? AND is_expired = FALSE AND updated_at >= datetime('now', 'localtime', '-50 minutes')
		)
	`, email, otp).Scan(&exists)

	if err != nil {
		log.Printf("Failed to verify OTP: %v\n", err)
		return false
	}

	return exists
}

func UpdateOtpStatus(email string) error {
	// Update the OTP status to expired
	_, err := db.Exec(`
	UPDATE otps SET is_expired = TRUE WHERE email = ?
	`, email)

	if err != nil {
		return fmt.Errorf("failed to update OTP status: %w", err)
	}

	return nil
}