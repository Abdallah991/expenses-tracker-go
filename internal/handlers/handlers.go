package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	// ! its not connecting
	connStr := os.Getenv("DATABASE_URL")

	if connStr == "" {
		// Log a fatal error or panic if the connection string is missing
		fmt.Println("FATAL: DATABASE_URL environment variable not set.")
		// In a real app, you might want to use log.Fatal(err) here
	}

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err) // Panic if connection setup fails
	}

	// Ping to verify the connection is alive
	if err = DB.Ping(); err != nil {
		panic(err) // Panic if the initial connection is bad
	}

	fmt.Println("✅ Successfully connected to the PostgreSQL database!")
}

// GetTransactionsHandler fetches all transactions from the database.
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure it's a GET request
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "Method not allowed. Only GET is supported."}`))
		return
	}

	// Query the database
	rows, err := DB.Query("SELECT id, amount FROM transaction ORDER BY id DESC")
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
		if err := rows.Scan(&t.ID, &t.Amount); err != nil {
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
