package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	// import handlers package
	"expenses-tracker-go/internal/auth"
	handlers "expenses-tracker-go/internal/handlers"
	"expenses-tracker-go/internal/ratelimit"
)

const (
	port = ":8080"
)

func main() {

	// initialize database connection
	handlers.InitDB()

	// initialize JWT configuration
	if err := auth.InitJWT(); err != nil {
		log.Fatalf("❌ Failed to initialize JWT: %v\n", err)
	}

	// initialize email service
	if err := handlers.InitEmailService(); err != nil {
		log.Fatalf("❌ Failed to initialize email service: %v\n", err)
	}

	// Setup the Router/Multiplexer**
	mux := http.NewServeMux()

	// * Public Routes (no authentication required)
	// Health check endpoint
	mux.HandleFunc("/status", handlers.StatusHandler)

	// Authentication endpoints with rate limiting
	mux.Handle("/auth/register", ratelimit.RegistrationRateLimitMiddleware()(http.HandlerFunc(handlers.RegisterHandler)))
	mux.Handle("/auth/login", ratelimit.LoginRateLimitMiddleware()(http.HandlerFunc(handlers.LoginHandler)))
	mux.HandleFunc("/auth/verify-email", handlers.VerifyEmailHandler)
	mux.Handle("/auth/resend-verification", ratelimit.VerificationResendRateLimitMiddleware()(http.HandlerFunc(handlers.ResendVerificationHandler)))
	mux.Handle("/auth/forgot-password", ratelimit.PasswordResetRateLimitMiddleware()(http.HandlerFunc(handlers.ForgotPasswordHandler)))
	mux.HandleFunc("/auth/reset-password", handlers.ResetPasswordHandler)
	mux.HandleFunc("/auth/refresh", handlers.RefreshTokenHandler)

	// * Protected Routes (authentication required)
	// Logout endpoint
	mux.Handle("/auth/logout", auth.RequireAuth(handlers.DB)(http.HandlerFunc(handlers.LogoutHandler)))

	// Transaction endpoints (protected)
	mux.Handle("/transactions", auth.RequireAuth(handlers.DB)(http.HandlerFunc(handlers.GetTransactionsHandler)))
	mux.Handle("/transaction", auth.RequireAuth(handlers.DB)(http.HandlerFunc(handlers.CreateTransactionHandler)))

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
	fmt.Printf("🔐 Authentication system initialized\n")
	fmt.Printf("📧 Email service initialized\n")

	// log.Fatal is used here because if the server fails to start, we want the application to exit.
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Could not listen on %s: %v\n", port, err)
	}
}
