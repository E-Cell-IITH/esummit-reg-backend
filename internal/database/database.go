package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
	"reg/internal/model"
)

var (
	dburl = os.Getenv("BLUEPRINT_DB_URL")
	db    *sql.DB
)

func New() {
	if db != nil {
		log.Println("Database already initialized")
		return
	}

	// Initialize the database connection
	dbConnection, err := sql.Open("sqlite3", dburl)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}

	// Test the connection
	if err := dbConnection.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Assign the connection and run migrations
	db = dbConnection
	if err := Migrate(); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database successfully initialized")
}

// Health checks the database health and returns health statistics.
func Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Check if the database connection is alive
	err := db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("Database unreachable: %v", err)
		log.Printf("Database health check failed: %v", err)
		return stats
	}

	// Collect database statistics
	stats["status"] = "up"
	stats["message"] = "Database is healthy"

	dbStats := db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	return stats
}

// Close terminates the database connection.
func Close() error {
	if db != nil {
		log.Printf("Closing database connection: %s", dburl)
		return db.Close()
	}
	log.Println("Database connection is already closed or not initialized")
	return nil
}

// Migrate creates the required tables in the database.
func Migrate() error {
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	// Define the schema creation queries
	createRegistrationsTableQuery := `
    CREATE TABLE IF NOT EXISTS registrations (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        sname TEXT,
        fname TEXT,
        pocname TEXT,
        contact TEXT,
        startup TEXT,
        service TEXT,
        email TEXT,
        semail TEXT,
        ifocus TEXT,
        ayears TEXT,
        location TEXT,
        city TEXT,
        about TEXT
    );
    `
	createOtpsTableQuery := `
    CREATE TABLE IF NOT EXISTS otps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE, 
        otp TEXT NOT NULL,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `

	// Execute the queries
	_, err := db.Exec(createRegistrationsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create registrations table: %w", err)
	}

	_, err = db.Exec(createOtpsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create otps table: %w", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}

// CreateRegistration inserts a new registration into the database.
func CreateRegistration(ctx context.Context, data model.RegistrationData) (int64, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection is not initialized")
	}

	query := `
    INSERT INTO registrations (
        sname, fname, pocname, contact, startup,
        service, email, semail, ifocus, ayears,
        location, city, about
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
	result, err := db.ExecContext(ctx, query,
		data.SName, data.FName, data.POCName, data.Contact, data.Startup,
		data.Service, data.Email, data.SEmail, data.IFocus, data.AYears,
		data.Location, data.City, data.About,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert registration: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert ID: %w", err)
	}

	return id, nil
}
