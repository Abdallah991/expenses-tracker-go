package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
)

// RequireAuth is middleware that requires a valid JWT token
func RequireAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			tokenString, err := ExtractTokenFromHeader(authHeader)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", err.Error())
				return
			}

			// Validate the token
			claims, err := ValidateAccessToken(tokenString)
			if err != nil {
				writeErrorResponse(w, http.StatusUnauthorized, "Invalid token", err.Error())
				return
			}

			// Check if user still exists and is active
			var userExists bool
			err = db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND email_verified = true)",
				claims.UserID,
			).Scan(&userExists)
			if err != nil {
				writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to verify user")
				return
			}

			if !userExists {
				writeErrorResponse(w, http.StatusUnauthorized, "User not found or not verified", "")
				return
			}

			// Add user information to request context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (int, error) {
	userID, ok := ctx.Value(UserIDKey).(int)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// GetUserEmailFromContext extracts user email from request context
func GetUserEmailFromContext(ctx context.Context) (string, error) {
	email, ok := ctx.Value(UserEmailKey).(string)
	if !ok {
		return "", fmt.Errorf("user email not found in context")
	}
	return email, nil
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// writeErrorResponse writes a JSON error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, error, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Details: details,
	}

	json.NewEncoder(w).Encode(response)
}

// OptionalAuth is middleware that validates JWT token if present but doesn't require it
func OptionalAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			tokenString, err := ExtractTokenFromHeader(authHeader)
			if err != nil {
				// Invalid header format, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Validate the token
			claims, err := ValidateAccessToken(tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Check if user still exists and is active
			var userExists bool
			err = db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND email_verified = true)",
				claims.UserID,
			).Scan(&userExists)
			if err != nil || !userExists {
				// User not found or not verified, continue without authentication
				next.ServeHTTP(w, r)
				return
			}

			// Add user information to request context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			// Call the next handler with the updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
