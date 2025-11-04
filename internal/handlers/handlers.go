package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	// add these specific imports
	"expenses-tracker-go/internal/auth"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

// Global database connection pool (initialize in main)
var DB *sql.DB

// ? StatusHandler is the HTTP handler for the /status endpoint.
// indicating the application is live.
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	//Check the HTTP Method, if not GET, return error
	if r.Method != http.MethodGet {
		// * status error 405, wrong method
		w.WriteHeader(http.StatusMethodNotAllowed)
		// Optionally write a JSON response for clarity
		w.Write([]byte(`{"error": "Method not allowed. Only GET is supported."}`))
		return
	}
	// Set the response header to indicate JSON content
	w.Header().Set("Content-Type", "application/json")

	//*  Set a 200 OK status code
	w.WriteHeader(http.StatusOK)

	// Create and encode the response
	response := map[string]string{
		"status":      "live",
		"application": "Go Simple Web Server",
		"message":     "Application is live and running!",
	}
	json.NewEncoder(w).Encode(response)
}

// ? initDB establishes the database connection. Call this from main.go
func InitDB() {
	// Load values from .env file into environment variables
	e := godotenv.Load()
	if e != nil {
		// This is usually fine if running in a non-dev environment
		// that uses real environment vars.
		fmt.Println("Could not load .env file:", e)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set.")
	}

	fmt.Println("Connecting to database...")

	// Parse connection string using pgx
	config, err := pgx.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Failed to parse connection string: %v", err)
	}

	// Force IPv4 resolution if hostname is used (for Render compatibility)
	if config.Host != "" {
		// Check if it's a hostname (not already an IP address)
		if net.ParseIP(config.Host) == nil {
			// It's a hostname, try to resolve to IPv4
			ipAddr, resolveErr := net.ResolveIPAddr("ip4", config.Host)
			if resolveErr == nil {
				fmt.Printf("Resolved %s to IPv4: %s\n", config.Host, ipAddr.IP.String())
				config.Host = ipAddr.IP.String()
			} else {
				// If IPv4-only resolution fails, try regular lookup
				fmt.Printf("IPv4-only resolution failed, trying regular lookup: %v\n", resolveErr)
				ips, lookupErr := net.LookupIP(config.Host)
				if lookupErr == nil {
					// Find IPv4 address
					for _, ip := range ips {
						if ip.To4() != nil {
							fmt.Printf("Resolved %s to IPv4: %s\n", config.Host, ip.To4().String())
							config.Host = ip.To4().String()
							break
						}
					}
				}
			}
		}
	}

	// Convert pgx config back to connection string for stdlib
	connStr = stdlib.RegisterConnConfig(config)

	var dbErr error
	DB, dbErr = sql.Open("pgx", connStr)
	if dbErr != nil {
		log.Fatalf("Failed to open database connection: %v", dbErr)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Ping with retry logic
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		if err = DB.Ping(); err == nil {
			fmt.Println("✅ Successfully connected to the PostgreSQL database!")
			return
		}
		fmt.Printf("Database ping attempt %d/%d failed: %v\n", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}

	log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
}

// GetTransactionsHandler fetches all transactions from the database for the authenticated user.
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a GET request
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed. Only GET is supported."}`))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Query the database for user's transactions only
	rows, err := DB.Query("SELECT id, amount, user_id FROM transaction WHERE user_id = $1 ORDER BY id DESC", userID)
	if err != nil {
		http.Error(w, "Database query failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close() // Ensure rows are closed after the function returns

	// 2. Collect results
	transactions := []Transaction{}
	for rows.Next() {
		var t Transaction
		// Scan the values from the current row into the Transaction struct
		if err := rows.Scan(&t.ID, &t.Amount, &t.UserID); err != nil {
			http.Error(w, "Error scanning transaction row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		transactions = append(transactions, t)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, "Error iterating over transactions: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Respond with JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

// CreateTransactionHandler handles POST requests to insert a new transaction for the authenticated user.
func CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Check HTTP Method
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed. Only POST is supported."}`))
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// 2. Decode the Request Body
	var t Transaction
	// Ensure we only read a reasonable amount of data to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, 1048576) // 1MB limit

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&t); err != nil {
		http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Since 'id' is SERIAL (auto-generated) in the database, we only use 'Amount' for insert.
	// Ensure amount is non-zero (simple validation)
	if t.Amount == 0 {
		http.Error(w, "Transaction amount cannot be zero.", http.StatusBadRequest)
		return
	}

	// 3. Execute INSERT Query
	// The RETURNING id clause is crucial to get the auto-generated ID back
	sqlStatement := `
        INSERT INTO transaction (amount, user_id)
        VALUES ($1, $2)
        RETURNING id`

	// Use DB.QueryRow for single row return (the new ID)
	err = DB.QueryRow(sqlStatement, t.Amount, userID).Scan(&t.ID)

	if err != nil {
		// Log the detailed error (for server logs) but return a generic 500
		fmt.Printf("Error inserting transaction: %v\n", err)
		http.Error(w, "Failed to insert transaction into database.", http.StatusInternalServerError)
		return
	}

	// Set the user ID in the response
	t.UserID = userID

	// 4. Respond with the Created Transaction
	w.Header().Set("Content-Type", "application/json")
	// Use 201 Created status code for successful resource creation
	w.WriteHeader(http.StatusCreated)

	// Respond with the transaction, now including the auto-generated ID and user ID
	json.NewEncoder(w).Encode(t)
}
