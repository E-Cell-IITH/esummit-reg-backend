package database

func InitiatePayment(amount float64, userId int) (int64, error) {
	result, err := db.Exec(`INSERT INTO payments_initiate (amount, user_id) VALUES (?, ?)`, amount, userId)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func CreatePaymentRecord(txnId string, userID int, amount float64) (int64, error) {
	result, err := db.Exec(`INSERT INTO transactions (id, user_id, amount) VALUES (?, ?, ?)`, txnId, userID, amount)
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