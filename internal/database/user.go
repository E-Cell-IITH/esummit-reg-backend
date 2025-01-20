package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reg/internal/model"
)

func CreateUser(ctx context.Context, user model.User) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	dataJSON, err := json.Marshal(user.Data)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal user data: %w", err)
	}

	query := `
    INSERT INTO users (email, name, contact_number, data)
    VALUES (?, ?, ?, ?)
    `
	result, err := db.ExecContext(ctx, query, user.Email, user.Name, user.ContactNumber, string(dataJSON))
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return id, nil
}
func UserExists(email string) bool {
	// Check if the user exists in the "users" table
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to check if user exists: %v\n", err)
		return false
	}

	return exists
}

func UserExistsByID(id string) bool {
	// Check if the user exists in the "users" table
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", id).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		log.Printf("Failed to check if user exists: %v\n", err)
		return false
	}

	return exists
}

func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
    SELECT id, email, name, contact_number, data
    FROM users
    WHERE email = ?
    `
	row := db.QueryRowContext(ctx, query, email)

	var user model.User
	var dataJSON string
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.ContactNumber, &dataJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	err = json.Unmarshal([]byte(dataJSON), &user.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

func GetUserById(ctx context.Context, id int64) (*model.User, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
	SELECT email, name, contact_number
	FROM users
	WHERE id = ?
	`
	row := db.QueryRowContext(ctx, query, id)

	var user model.User
	err := row.Scan(&user.Email, &user.Name, &user.ContactNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &user, nil
}

func GetMeUser(ctx context.Context, id int64) (*model.User, int, error) {
	if db == nil {
		return nil, -1, fmt.Errorf("database connection is not initialized")
	}

	query := `
	SELECT 
    	u.name,
    	u.email,
    	u.contact_number,
    	COALESCE(pt.ticket_id, '-1') AS ticket_id
	FROM 
    	users u
	LEFT JOIN 
    	purchased_tickets pt ON u.id = pt.user_id 
	WHERE u.id = ?;
	`
	row := db.QueryRowContext(ctx, query, id)

	var user model.User
	var ticketID int
	err := row.Scan(&user.Email, &user.Name, &user.ContactNumber, &ticketID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, -1, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, -1, fmt.Errorf("failed to fetch user: %w", err)
	}

	return &user, ticketID, nil
}
