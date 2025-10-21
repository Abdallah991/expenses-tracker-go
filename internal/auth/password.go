package auth

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const (
	// Password requirements
	MinPasswordLength = 8
	MaxPasswordLength = 128 //! is this allowed ? considering that the bycrypt library has a max length of 72 bytes
	BcryptCost        = 12
)

var (
	// Password validation regex patterns
	hasUppercase = regexp.MustCompile(`[A-Z]`)
	hasLowercase = regexp.MustCompile(`[a-z]`)
	hasNumber    = regexp.MustCompile(`[0-9]`)
	hasSpecial   = regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?~` + "`" + `]`)
)

// PasswordValidationError represents password validation errors
type PasswordValidationError struct {
	Message string
}

func (e *PasswordValidationError) Error() string {
	return e.Message
}

// HashPassword hashes a password using bcrypt with the configured cost
func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		// ! there is three kinds of errors here we need to consider
		return "", err
	}

	return string(hashedBytes), nil
}

// ComparePassword - compares a plain text password with its equivalant hashed password
func ComparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// ValidatePassword - validates password strength according to security requirements
func ValidatePassword(password string) error {
	// length - short scenario
	if len(password) < MinPasswordLength {
		return &PasswordValidationError{
			Message: "Password must be at least 8 characters long",
		}
	}
	// length - long scenario
	if len(password) > MaxPasswordLength {
		return &PasswordValidationError{
			Message: "Password must be no more than 128 characters long",
		}
	}

	// Check for common weak passwords - //? add more weak passwords
	weakPasswords := []string{
		"password", "123456", "123456789", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome", "monkey",
	}

	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if strings.Contains(lowerPassword, weak) {
			return &PasswordValidationError{
				Message: "Password contains common weak patterns",
			}
		}
	}

	// Check character requirements - uppercase
	if !hasUppercase.MatchString(password) {
		return &PasswordValidationError{
			Message: "Password must contain at least one uppercase letter",
		}
	}
	// lowercase
	if !hasLowercase.MatchString(password) {
		return &PasswordValidationError{
			Message: "Password must contain at least one lowercase letter",
		}
	}
	// numeric
	if !hasNumber.MatchString(password) {
		return &PasswordValidationError{
			Message: "Password must contain at least one number",
		}
	}
	// special character
	if !hasSpecial.MatchString(password) {
		return &PasswordValidationError{
			Message: "Password must contain at least one special character",
		}
	}

	// Check for repeated characters (more than 3 in a row)
	if hasRepeatedChars(password) {
		return &PasswordValidationError{
			Message: "Password cannot contain more than 3 repeated characters in a row",
		}
	}

	return nil
}

// hasRepeatedChars checks if password has more than 3 repeated characters
func hasRepeatedChars(password string) bool {
	// if length less than 4 return false
	if len(password) < 4 {
		return false
	}
	// return true when it condition applies, false otherwise
	for i := 0; i < len(password)-3; i++ {
		if password[i] == password[i+1] && password[i+1] == password[i+2] && password[i+2] == password[i+3] {
			return true
		}
	}

	return false
}

// IsPasswordValidationError checks if an error is a password validation error
func IsPasswordValidationError(err error) bool {
	var validationErr *PasswordValidationError
	return errors.As(err, &validationErr)
}
