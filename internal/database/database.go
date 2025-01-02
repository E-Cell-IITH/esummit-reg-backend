package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"reg/internal/model"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)


type Service interface {
    // Health returns a map of health status information.
    Health() map[string]string

    // Close terminates the database connection.
    Close() error

    // Migrate handles creating/updating your database schema.
    Migrate() error

    // CreateRegistration inserts a new record into your registrations table.
    // Returns the newly created record’s ID and an error if something goes wrong.
    CreateRegistration(ctx context.Context, data model.RegistrationData) (int64, error)

}

type service struct {
    db *sql.DB
}

var (
    dburl      = os.Getenv("BLUEPRINT_DB_URL")
    dbInstance *service
)

// New returns a singleton service (i.e., reuses the DB connection if it exists).
func New() Service {
    // Reuse the connection if already instantiated
    if dbInstance != nil {
        return dbInstance
    }

    // Connect to the SQLite database
    db, err := sql.Open("sqlite3", dburl)
    if err != nil {
        // This will not be a connection error, but a DSN parse error or
        // another initialization error.
        log.Fatal(err)
    }

    dbInstance = &service{
        db: db,
    }
    return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
    defer cancel()

    stats := make(map[string]string)

    // Ping the database
    err := s.db.PingContext(ctx)
    if err != nil {
        stats["status"] = "down"
        stats["error"] = fmt.Sprintf("db down: %v", err)
        log.Fatalf("db down: %v", err) // Log the error and terminate the program
        return stats
    }

    // Database is up, add more statistics
    stats["status"] = "up"
    stats["message"] = "It's healthy"

    // Get database stats (like open connections, in use, idle, etc.)
    dbStats := s.db.Stats()
    stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
    stats["in_use"] = strconv.Itoa(dbStats.InUse)
    stats["idle"] = strconv.Itoa(dbStats.Idle)
    stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
    stats["wait_duration"] = dbStats.WaitDuration.String()
    stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
    stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

    // Evaluate stats to provide a health message
    if dbStats.OpenConnections > 40 { // Arbitrary threshold example
        stats["message"] = "The database is experiencing heavy load."
    }

    if dbStats.WaitCount > 1000 {
        stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
    }

    return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
    log.Printf("Disconnected from database: %s", dburl)
    return s.db.Close()
}

// Migrate creates (or updates) your database schema.
// You can add more CREATE TABLE statements or run migrations as needed.
func (s *service) Migrate() error {
    createTableQuery := `
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
    _, err := s.db.Exec(createTableQuery)
    if err != nil {
        return fmt.Errorf("failed to create or verify registrations table: %w", err)
    }
    return nil
}

// CreateRegistration inserts a new record into the "registrations" table.
// Returns the ID of the newly inserted record and an error if something goes wrong.
func (s *service) CreateRegistration(ctx context.Context, data model.RegistrationData) (int64, error) {
    query := `
        INSERT INTO registrations (
            sname, fname, pocname, contact, startup,
            service, email, semail, ifocus, ayears,
            location, city, about
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `
    result, err := s.db.ExecContext(ctx, query,
        data.SName,
        data.FName,
        data.POCName,
        data.Contact,
        data.Startup,
        data.Service,
        data.Email,
        data.SEmail,
        data.IFocus,
        data.AYears,
        data.Location,
        data.City,
        data.About,
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
