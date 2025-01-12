package database

import "fmt"

// SaveOtp saves or updates an OTP in the database for a given email.
func SaveOtp(email string, otp string) error {
	// Save OTP in the "otps" table
	_, err := db.Exec(`
		INSERT INTO otps (email, otp, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(email) DO UPDATE SET 
			otp = excluded.otp, 
			updated_at = CURRENT_TIMESTAMP
	`, email, otp)

	if err != nil {
		return fmt.Errorf("failed to save OTP: %w", err)
	}

	return nil
}
