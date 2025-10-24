package handlers

import (
	"database/sql"
	"encoding/json"
	"expenses-tracker-go/internal/auth"
	"expenses-tracker-go/internal/email"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

// Global email service instance
var emailService *email.EmailService

// InitEmailService initializes the email service
func InitEmailService() error {
	var err error
	emailService, err = email.NewEmailService()
	return err
}

// isValidEmail validates email format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", "")
		return
	}

	// Validate password strength
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Password validation failed", err.Error())
		return
	}

	// Check if email already exists
	var existingUserID int
	err := DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingUserID)
	if err == nil {
		writeErrorResponse(w, http.StatusConflict, "Email already registered", "")
		return
	} else if err != sql.ErrNoRows {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to check email")
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Password hashing failed", err.Error())
		return
	}

	// Generate verification token
	verificationToken, err := auth.GenerateSecureToken(32)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	verificationExpires := time.Now().Add(24 * time.Hour)

	// Create user
	var userID int
	err = DB.QueryRow(`
		INSERT INTO users (email, password_hash, verification_token, verification_token_expires)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, req.Email, hashedPassword, verificationToken, verificationExpires).Scan(&userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err.Error())
		return
	}

	// Send verification email
	if err := emailService.SendVerificationEmail(req.Email, verificationToken); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Failed to send verification email: %v\n", err)
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "User registered successfully. Please check your email to verify your account.",
	})
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", "")
		return
	}

	// Get user from database
	var user User
	err := DB.QueryRow(`
		SELECT id, email, password_hash, email_verified, failed_login_attempts, locked_until
		FROM users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.EmailVerified,
		&user.FailedLoginAttempts, &user.LockedUntil,
	)

	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to fetch user")
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		writeErrorResponse(w, http.StatusLocked, "Account locked", "Too many failed login attempts. Please try again later.")
		return
	}

	// Check if email is verified
	if !user.EmailVerified {
		writeErrorResponse(w, http.StatusForbidden, "Email not verified", "Please verify your email before logging in")
		return
	}

	// Verify password
	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		// Increment failed login attempts
		newAttempts := user.FailedLoginAttempts + 1
		var lockedUntil *time.Time

		if newAttempts >= 5 {
			lockTime := time.Now().Add(15 * time.Minute)
			lockedUntil = &lockTime
		}

		DB.Exec(`
			UPDATE users 
			SET failed_login_attempts = $1, locked_until = $2, updated_at = NOW()
			WHERE id = $3
		`, newAttempts, lockedUntil, user.ID)

		writeErrorResponse(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// Reset failed login attempts on successful login
	DB.Exec(`
		UPDATE users 
		SET failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
		WHERE id = $1
	`, user.ID)

	// Generate tokens
	accessToken, err := auth.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Refresh token generation failed", err.Error())
		return
	}

	// Store refresh token in database
	refreshExpires := time.Now().Add(auth.GetRefreshTokenExpiry())
	_, err = DB.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, user.ID, refreshToken, refreshExpires)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to store refresh token", err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{
		Message:      "Login successful",
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.GetTokenExpiry().Seconds()),
		User:         user.ToUserResponse(),
	})
}

// VerifyEmailHandler handles email verification
func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only GET is supported")
		return
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Token required", "Please provide a verification token")
		return
	}

	// Verify token and update user
	var userID int
	err := DB.QueryRow(`
		SELECT id FROM users 
		WHERE verification_token = $1 
		AND verification_token_expires > NOW()
	`, token).Scan(&userID)

	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid or expired token", "")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to verify token")
		return
	}

	// Mark email as verified and clear verification token
	_, err = DB.Exec(`
		UPDATE users 
		SET email_verified = true, verification_token = NULL, verification_token_expires = NULL, updated_at = NOW()
		WHERE id = $1
	`, userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to verify email", err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Email verified successfully. You can now log in.",
	})
}

// ResendVerificationHandler handles resending verification emails
func ResendVerificationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", "")
		return
	}

	// Check if user exists and is not verified
	var userID int
	var emailVerified bool
	err := DB.QueryRow(`
		SELECT id, email_verified FROM users WHERE email = $1
	`, req.Email).Scan(&userID, &emailVerified)

	if err == sql.ErrNoRows {
		// Don't reveal if email exists or not
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse{
			Message: "If the email exists and is not verified, a verification email has been sent.",
		})
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to check user")
		return
	}

	if emailVerified {
		writeErrorResponse(w, http.StatusBadRequest, "Email already verified", "")
		return
	}

	// Generate new verification token
	verificationToken, err := auth.GenerateSecureToken(32)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	verificationExpires := time.Now().Add(24 * time.Hour)

	// Update verification token
	_, err = DB.Exec(`
		UPDATE users 
		SET verification_token = $1, verification_token_expires = $2, updated_at = NOW()
		WHERE id = $3
	`, verificationToken, verificationExpires, userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update verification token", err.Error())
		return
	}

	// Send verification email
	if err := emailService.SendVerificationEmail(req.Email, verificationToken); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to send verification email", err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Verification email sent successfully.",
	})
}

// ForgotPasswordHandler handles forgot password requests
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate email format
	if !isValidEmail(req.Email) {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid email format", "")
		return
	}

	// Check if user exists
	var userID int
	err := DB.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&userID)
	if err == sql.ErrNoRows {
		// Don't reveal if email exists or not
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SuccessResponse{
			Message: "If the email exists, a password reset link has been sent.",
		})
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to check user")
		return
	}

	// Generate reset token
	resetToken, err := auth.GenerateSecureToken(32)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	resetExpires := time.Now().Add(1 * time.Hour)

	// Update reset token
	_, err = DB.Exec(`
		UPDATE users 
		SET reset_token = $1, reset_token_expires = $2, updated_at = NOW()
		WHERE id = $3
	`, resetToken, resetExpires, userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update reset token", err.Error())
		return
	}

	// Send reset email
	if err := emailService.SendPasswordResetEmail(req.Email, resetToken); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to send reset email", err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "If the email exists, a password reset link has been sent.",
	})
}

// ResetPasswordHandler handles password reset
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate new password strength
	if err := auth.ValidatePassword(req.NewPassword); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Password validation failed", err.Error())
		return
	}

	// Verify reset token
	var userID int
	err := DB.QueryRow(`
		SELECT id FROM users 
		WHERE reset_token = $1 
		AND reset_token_expires > NOW()
	`, req.Token).Scan(&userID)

	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid or expired reset token", "")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to verify reset token")
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Password hashing failed", err.Error())
		return
	}

	// Update password and clear reset token
	_, err = DB.Exec(`
		UPDATE users 
		SET password_hash = $1, reset_token = NULL, reset_token_expires = NULL, 
		    failed_login_attempts = 0, locked_until = NULL, updated_at = NOW()
		WHERE id = $2
	`, hashedPassword, userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update password", err.Error())
		return
	}

	// Invalidate all refresh tokens for this user
	DB.Exec("DELETE FROM refresh_tokens WHERE user_id = $1", userID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Password reset successfully. Please log in with your new password.",
	})
}

// RefreshTokenHandler handles token refresh
func RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Verify refresh token
	var userID int
	var email string
	err := DB.QueryRow(`
		SELECT u.id, u.email 
		FROM users u
		JOIN refresh_tokens rt ON u.id = rt.user_id
		WHERE rt.token = $1 
		AND rt.expires_at > NOW()
		AND u.email_verified = true
	`, req.RefreshToken).Scan(&userID, &email)

	if err == sql.ErrNoRows {
		writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired refresh token", "")
		return
	} else if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Database error", "Failed to verify refresh token")
		return
	}

	// Generate new access token
	accessToken, err := auth.GenerateAccessToken(userID, email)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Token generation failed", err.Error())
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RefreshTokenResponse{
		Message:     "Token refreshed successfully",
		AccessToken: accessToken,
		ExpiresIn:   int64(auth.GetTokenExpiry().Seconds()),
	})
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, http.StatusMethodNotAllowed, "Method not allowed", "Only POST is supported")
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", "")
		return
	}

	// Get refresh token from request body
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Delete the specific refresh token
	result, err := DB.Exec(`
		DELETE FROM refresh_tokens 
		WHERE token = $1 AND user_id = $2
	`, req.RefreshToken, userID)

	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to logout", err.Error())
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid refresh token", "")
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{
		Message: "Logged out successfully",
	})
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
