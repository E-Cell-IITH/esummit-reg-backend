package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func InitiatePayment(amount float64, userId int) (int64, error) {
	result, err := db.Exec(`INSERT INTO payments_initiate (amount, user_id) VALUES (?, ?)`, amount, userId)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func CreatePaymentRecord(txnId string, userID int, amount float64, ticketTitle string, isAccommodation bool) (int64, error) {
	// First, check if a record with the same txnId already exists
	var exists int
	err := db.QueryRow(`SELECT 1 FROM transactions WHERE id = ?`, txnId).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}
	if exists == 1 {
		// If the txnId already exists, return -1
		return -1, nil
	}

	result, err := db.Exec(`INSERT INTO transactions (id, user_id, amount, ticket_title, isAccommodation) VALUES (?, ?, ?, ?, ?)`, txnId, userID, amount, ticketTitle, isAccommodation)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func GetCountOfOrders() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func AddSuccessfulTxnIds(txnId string, amount float64) (int, int, error) {
	fmt.Println("txnId: ", txnId)
	qCheck := `SELECT 1 FROM transactions WHERE id = ?;`

	qAlreadyVerified := `SELECT 1 FROM transactions WHERE id = ? AND is_verified = TRUE;`

	qUpdate := `UPDATE transactions
                SET is_verified = TRUE
                WHERE id = ?;`

	qSelect := `SELECT 
                    CASE 
                        WHEN EXISTS (SELECT 1 FROM transactions WHERE id = ? AND is_verified = TRUE) THEN 1
                        ELSE -1
                    END AS result,
                    user_id
                FROM transactions
                WHERE id = ?;`

	var exists int
	err := db.QueryRow(qCheck, txnId).Scan(&exists)
	if err != nil || exists == 0 {
		fmt.Println(err)
		return -1, 0, nil
	}

	var alreadyVerified int
	err = db.QueryRow(qAlreadyVerified, txnId).Scan(&alreadyVerified)

	if err != nil {
		if err == sql.ErrNoRows {
			alreadyVerified = 0
		} else {
			return -1, 0, fmt.Errorf("select error while querying qAlreadyVerified: %w", err)
		}
	}

	if alreadyVerified == 1 {
		return 45, 0, nil
	}

	_, err = db.Exec(qUpdate, txnId)
	if err != nil {
		return -1, 0, fmt.Errorf("update error: %w", err)
	}

	var result, userId int
	err = db.QueryRow(qSelect, txnId, txnId).Scan(&result, &userId)
	if err != nil {
		return -1, 0, fmt.Errorf("select error while querying qSelect: %w", err)
	}

	return result, userId, nil
}

func AddTickets(userID int, txnID string) (string, error) {
	var (
		ticketTitle     string
		price           float64
		isAccommodation bool
	)

	query := `
		SELECT ticket_title, amount, isAccommodation
		FROM transactions
		WHERE id = ? AND user_id = ? AND is_verified = TRUE
	`
	err := db.QueryRow(query, txnID, userID).Scan(&ticketTitle, &price, &isAccommodation)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ticketTitle,  fmt.Errorf("transaction not found or not verified")
		}
		return ticketTitle, err
	}

	insertQuery := `
		INSERT INTO purchased_tickets (user_id, ticket_title, price, isAccommodation)
		VALUES (?, ?, ?, ?)
	`
	_, err = db.Exec(insertQuery, userID, ticketTitle, price, isAccommodation)
	if err != nil {
		return ticketTitle, fmt.Errorf("failed to add ticket: %v", err)
	}

	log.Printf("Ticket successfully added for user %d with transaction ID %s", userID, txnID)
	return ticketTitle, nil
}

func AddBasicTickets(userID int, ticketTitle string) error {
	insertQuery := `
		INSERT INTO purchased_tickets (user_id, ticket_title, price, isAccommodation)
		VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(insertQuery, userID, ticketTitle, -1, false)
	if err != nil {
		return fmt.Errorf("failed to add ticket: %v", err)
	}

	log.Printf("Ticket successfully added for user %d", userID)
	return nil
}
