package database

func CreateOrder(userID int64, razorpayOrderID string, amount float64, receipt string) (int64, error) {
	result, err := db.Exec(`INSERT INTO orders (razorpay_order_id, user_id, amount, receipt) VALUES (?, ?, ?, ?)`,
		razorpayOrderID, userID, amount, receipt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func RecordPayment(razorpayPaymentID, razorpayOrderID string, userID int, amount float64, status string) (int64, error) {
    result, err := db.Exec(`INSERT INTO payments (razorpay_payment_id, razorpay_order_id, amount, status, user_id) VALUES (?, ?, ?, ?, ?)`,
        razorpayPaymentID, razorpayOrderID, amount, status, userID)
    if err != nil {
        return 0, err
    }
    return result.LastInsertId()
}

func GetCountOfOrders() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}