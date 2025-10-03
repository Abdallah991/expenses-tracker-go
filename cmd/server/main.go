package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	// import handlers package
	handlers "expenses-tracker-go/internal/handlers"
)

const (
	port = ":8080"
)

func main() {

	// initialize database connection
	handlers.InitDB()

	// Setup the Router/Multiplexer**
	mux := http.NewServeMux()

	// * Define Routes
	// Register the StatusHandler function for GET requests to the /status path.
	mux.HandleFunc("/status", handlers.StatusHandler)
	// Register the transactions API route
	mux.HandleFunc("/transactions", handlers.GetTransactionsHandler)

	// Configure the Server
	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,   // Maximum duration for reading the entire request
		WriteTimeout: 10 * time.Second,  // Maximum duration before timing out writes
		IdleTimeout:  120 * time.Second, // Maximum amount of time to wait for the next request when keep-alives are enabled
	}

	// Start the Server
	fmt.Printf("✅ Starting server on port %s...\n", port)

	// log.Fatal is used here because if the server fails to start, we want the application to exit.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Could not listen on %s: %v\n", port, err)
	}
}
