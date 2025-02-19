package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"reg/internal/model"

	_ "github.com/joho/godotenv/autoload"
	_ "modernc.org/sqlite"
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
	dbConnection, err := sql.Open("sqlite", dburl)
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
	
	CREATE INDEX IF NOT EXISTS idx_registrations_email ON registrations(email);

	CREATE TABLE IF NOT EXISTS pushed_registrations (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	registration_id INTEGER NOT NULL,
    	pushed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	FOREIGN KEY (registration_id) REFERENCES registrations(id)
	);

    `
	createQuery := `
    CREATE TABLE IF NOT EXISTS otps (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        email TEXT NOT NULL UNIQUE, 
        otp TEXT NOT NULL,
		is_expired BOOLEAN DEFAULT FALSE,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

	CREATE INDEX IF NOT EXISTS idx_otps_email ON otps(email);

	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		contact_number TEXT NOT NULL,
		data json
	);

	CREATE TABLE IF NOT EXISTS payments_initiate (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL NOT NULL,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS transactions (
		id TEXT NOT NULL UNIQUE,
		amount REAL NOT NULL,
		user_id INTEGER NOT NULL,
		is_verified BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ticket_title TEXT NOT NULL,
		isAccommodation BOOLEAN DEFAULT FALSE,
		coupon TEXT DEFAULT "",
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (id)
	);

	CREATE TABLE IF NOT EXISTS purchased_tickets (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		ticket_title TEXT NOT NULL,
		price REAL NOT NULL,
		isAccommodation BOOLEAN DEFAULT FALSE,
		coupon TEXT DEFAULT "",
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS pushed_purchased_tickets (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	pushed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		ticket_id INTEGER NOT NULL,
    	FOREIGN KEY (ticket_id) REFERENCES purchased_tickets(id)
	);
	CREATE TABLE IF NOT EXISTS pushed_txn (
    	id INTEGER PRIMARY KEY AUTOINCREMENT,
    	pushed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		txn_id TEXT NOT NULL,
    	FOREIGN KEY (txn_id) REFERENCES transactions(id)
	);
	
	
	`

	// INSERT INTO tickets (name, description, price) VALUES
	// 	('STANDARD', 'All Speaker Sessions, Startup Fair, Food Carnival', -1),
	// 	('VALUE FOR MONEY', 'All Speaker Sessions, Startup Fair, Food Carniva, Fetching Fortune Spectator', 399),
	// 	('PREMIUM',  'All Speaker Sessions, Startup Fair, Food Carniva, Fetching Fortune Spectator, Networking Dinner, Accommodation, (2 Days 1 Night)', 999);
	// `

	// Execute the queries
	_, err := db.Exec(createRegistrationsTableQuery)
	if err != nil {
		return fmt.Errorf("failed to create registrations table: %w", err)
	}

	_, err = db.Exec(createQuery)
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

func GetRegistrationsYetToPush(ctx context.Context) ([]model.RegistrationData, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
	SELECT id, sname, fname, pocname, contact, startup, service, email, semail, ifocus, ayears, location, city, about
	FROM registrations
	WHERE id NOT IN (SELECT registration_id FROM pushed_registrations)
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query registrations: %w", err)
	}
	defer rows.Close()

	var registrations []model.RegistrationData
	for rows.Next() {
		var reg model.RegistrationData
		err := rows.Scan(
			&reg.Id,
			&reg.SName, &reg.FName, &reg.POCName, &reg.Contact, &reg.Startup,
			&reg.Service, &reg.Email, &reg.SEmail, &reg.IFocus, &reg.AYears,
			&reg.Location, &reg.City, &reg.About,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration: %w", err)
		}

		registrations = append(registrations, reg)
	}

	return registrations, nil
}

func GetPurchasedTickets(ctx context.Context) ([]model.PurchasedTicketWithUser, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `
	SELECT pt.id, pt.user_id, pt.ticket_title, pt.price, pt.isAccommodation, pt.coupon, u.email, u.name, u.contact_number, u.data
	FROM purchased_tickets pt
	JOIN users u ON pt.user_id = u.id
	WHERE pt.id NOT IN (SELECT ticket_id FROM pushed_purchased_tickets)
	`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query purchased tickets: %w", err)
	}
	defer rows.Close()

	var tickets []model.PurchasedTicketWithUser
	for rows.Next() {
		var ticket model.PurchasedTicketWithUser
		err := rows.Scan(
			&ticket.ID,
			&ticket.UserID,
			&ticket.TicketTitle,
			&ticket.Price,
			&ticket.IsAccommodation,
			&ticket.Coupon,
			&ticket.User.Email,
			&ticket.User.Name,
			&ticket.User.ContactNumber,
			&ticket.User.Data,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchased ticket: %w", err)
		}

		tickets = append(tickets, ticket)
	}

	return tickets, nil
}

func MarkTicketAsPushed(ctx context.Context, data []model.PurchasedTicketWithUser) error {
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	insertQuery := `
	INSERT INTO pushed_purchased_tickets (ticket_id)
	VALUES (?)
	`
	for _, ticket := range data {
		_, err := tx.ExecContext(ctx, insertQuery, ticket.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert pushed ticket: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func MarkRegistrationAsPushed(ctx context.Context, data []model.RegistrationData) error {
	if db == nil {
		return fmt.Errorf("database connection is not initialized")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	insertQuery := `
	INSERT INTO pushed_registrations (registration_id)
	VALUES (?)
	`
	for _, reg := range data {
		_, err := tx.ExecContext(ctx, insertQuery, reg.Id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert pushed registration: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetIDsForPasses() []model.UserTicket {
	// List of email addresses extracted from the logs
	emailedUsers := []string{
		"kaizeltech@gmail.com",
		"nrhitesh111@gmail.com",
		"gyaani2004@gmail.com",
		"Jaswanthkumar431@gmail.com",
		"mailforpraveen999@gmail.com",
		"sreeja.darlz@gmail.com",
		"tejavishnu2000@gmail.com",
		"deepak08925@gmail.com",
		"dasarirahulpatel.drp@gmail.com",
		"kartheekdama2004@gmail.com",
		"varshitha03@gmail.com",
		"manojkiranb98@gmail.com",
		"mukeshchowdar777@gmail.com",
		"ravitejareddy875@gmail.com",
		"shivakotagiri532@gmail.com",
		"kevinpaul468@gmail.com",
		"akshayabejgum05@gmail.com",
		"Kaizeltech@gmail.com",
		"nishnabandari@gmail.com",
		"arkalavarshitha37@gmail.com",
		"likhitharambha@gmail.com",
		"mechinenil@gmail.com",
		"divyareddyavula17@gmail.com",
		"noothisrimulya@gmail.com",
		"vijayintelli72@gmail.com",
		"hajrafatima1212@gmail.com",
		"imabdullah5978@gmail.com",
		"mdsaif6304@gmail.com",
		"vijaysuru620@gmail.com",
		"palthisaketh93@gmail.com",
		"rajnikita05@gmail.com",
		"pranavpolawar123@gmail.com",
		"dasarianjani1@gmail.com",
		"anjali.gadikhana16@gmail.com",
		"nagirimihira1960@gmail.com",
		"gorantlasreeja589@gmail.com",
		"rojaberi2005@gmail.com",
		"swathi8379t@gmail.com",
		"ravi2182003@gmail.com",
		"sathvikrepala30@gmail.com",
		"nalajalaabhi2004@gmail.com",
		"22r01a67g0@cmrithyderabad.edu.in",
		"charangoud3333@gmail.com",
		"varunsaivarma8@gmail.com",
		"shivanisweety102@gmail.com",
		"nenavathkalyan300@gmail.com",
		"patnammahesh75@gmail.com",
		"sarayusiri15@gmail.com",
		"22r01a67b2@gmail.com",
		"mehulagarwal11111@gmail.com",
		"chaturvediprabhu939@gmail.com",
	}

	// Construct the NOT IN clause for the SQL query
	notInClause := "'" + strings.Join(emailedUsers, "', '") + "'"

	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.email, pt.ticket_title 
		FROM purchased_tickets pt
		JOIN users u ON pt.user_id = u.id
		WHERE u.email NOT IN (%s)
		ORDER BY pt.id
	`, notInClause)

	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var userTickets []model.UserTicket

	for rows.Next() {
		var ut model.UserTicket
		if err := rows.Scan(&ut.ID, &ut.Name, &ut.Email, &ut.TicketTitle); err != nil {
			log.Fatal(err)
		}
		trimmedTitle := strings.ReplaceAll(strings.TrimSpace(ut.TicketTitle), " ", "_")
		ut.UID = fmt.Sprintf("%d_%s_%s_GUEST", ut.ID, strings.ToUpper(trimmedTitle), strings.ToLower(ut.Email))
		userTickets = append(userTickets, ut)
	}

	// Check for errors from iteration
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	return userTickets
}