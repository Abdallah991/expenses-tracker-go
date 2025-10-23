package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestValidatePassword_Length tests password length validation
func TestValidatePassword_Length(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid password with exactly 8 characters",
			password: "Pass123!",
			wantErr:  false,
		},
		{
			name:     "password too short - 7 characters",
			password: "Pass12!",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "password too short - 0 characters",
			password: "",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "password too short - 1 character",
			password: "a",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "password with exactly 128 characters",
			password: strings.Repeat("Pass123!", 16), // 8 * 16 = 128
			wantErr:  false,
		},
		{
			name:     "password too long - 129 characters",
			password: strings.Repeat("Pass123!", 16) + "a", // 128 + 1 = 129
			wantErr:  true,
			errMsg:   "no more than 128 characters",
		},
		{
			name:     "password too long - 200 characters",
			password: strings.Repeat("a", 200),
			wantErr:  true,
			errMsg:   "no more than 128 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePassword_CharacterRequirements tests character requirements
func TestValidatePassword_CharacterRequirements(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "missing uppercase letter",
			password: "mypass123!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "missing lowercase letter",
			password: "MYPASS123!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "missing number",
			password: "MyPass!@#",
			wantErr:  true,
			errMsg:   "number",
		},
		{
			name:     "missing special character",
			password: "MyPass123",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "only numbers",
			password: "98765432",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "only letters",
			password: "MyPassABC",
			wantErr:  true,
			errMsg:   "number",
		},
		{
			name:     "all requirements met",
			password: "Pass123!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePassword_WeakPasswords tests weak password detection
func TestValidatePassword_WeakPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "contains password",
			password: "MyPassword123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "contains 123456",
			password: "Test123456!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "contains qwerty",
			password: "Qwerty123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "contains admin",
			password: "Admin123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "case-insensitive check - PASSWORD",
			password: "PASSWORD123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "strong password",
			password: "SecureP@ss1",
			wantErr:  false,
		},
		{
			name:     "contains letmein",
			password: "Letmein123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
		{
			name:     "contains welcome",
			password: "Welcome123!",
			wantErr:  true,
			errMsg:   "weak patterns",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePassword_RepeatedCharacters tests repeated character detection
func TestValidatePassword_RepeatedCharacters(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "four repeated chars",
			password: "Pass1111!",
			wantErr:  true,
			errMsg:   "more than 3 repeated characters",
		},
		{
			name:     "five repeated chars",
			password: "Passaaaaa1!",
			wantErr:  true,
			errMsg:   "more than 3 repeated characters",
		},
		{
			name:     "repeated at start",
			password: "AAAApass123!",
			wantErr:  true,
			errMsg:   "more than 3 repeated characters",
		},
		{
			name:     "repeated at end",
			password: "Pass123!!!!",
			wantErr:  true,
			errMsg:   "more than 3 repeated characters",
		},
		{
			name:     "repeated in middle",
			password: "Pa1111ss!",
			wantErr:  true,
			errMsg:   "more than 3 repeated characters",
		},
		{
			name:     "three repeated chars (allowed)",
			password: "Pass111!",
			wantErr:  false,
		},
		{
			name:     "two repeated chars (allowed)",
			password: "Pass11!A",
			wantErr:  false,
		},
		{
			name:     "no repeated chars",
			password: "Pass123!",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestValidatePassword_EdgeCases tests edge cases
func TestValidatePassword_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "empty string",
			password: "",
			wantErr:  true,
			errMsg:   "at least 8 characters",
		},
		{
			name:     "whitespace only",
			password: "        ",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "unicode characters",
			password: "Pāss123!",
			wantErr:  false,
		},
		{
			name:     "all special chars included",
			password: "P@ss1#$%",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePassword() expected error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidatePassword() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidatePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestHashPassword tests password hashing
func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password returns hashed string",
			password: "SecureP@ss1",
			wantErr:  false,
		},
		{
			name:     "invalid password returns validation error",
			password: "short",
			wantErr:  true,
		},
		{
			name:     "too short password returns error",
			password: "Pass1!",
			wantErr:  true,
		},
		{
			name:     "too long password returns error",
			password: strings.Repeat("a", 200),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("HashPassword() unexpected error = %v", err)
				return
			}

			// Test hash format and properties
			if hash == "" {
				t.Errorf("HashPassword() returned empty hash")
			}

			// Test bcrypt format
			if !strings.HasPrefix(hash, "$2a$") {
				t.Errorf("HashPassword() hash doesn't start with bcrypt prefix, got %v", hash[:10])
			}

			// Test hash length (bcrypt hashes are 60 characters)
			if len(hash) != 60 {
				t.Errorf("HashPassword() hash length = %v, want 60", len(hash))
			}

			// Test cost factor
			if !strings.Contains(hash, "$12$") {
				t.Errorf("HashPassword() hash doesn't contain cost factor 12")
			}
		})
	}
}

// TestHashPassword_Salting tests that same password produces different hashes
func TestHashPassword_Salting(t *testing.T) {
	password := "SecureP@ss1"

	hash1, err1 := HashPassword(password)
	if err1 != nil {
		t.Fatalf("HashPassword() first attempt failed: %v", err1)
	}

	hash2, err2 := HashPassword(password)
	if err2 != nil {
		t.Fatalf("HashPassword() second attempt failed: %v", err2)
	}

	// Same password should produce different hashes due to salting
	if hash1 == hash2 {
		t.Errorf("HashPassword() same password produced identical hashes, salting not working")
	}

	// Both hashes should be valid bcrypt hashes
	if !strings.HasPrefix(hash1, "$2a$") || !strings.HasPrefix(hash2, "$2a$") {
		t.Errorf("HashPassword() produced invalid bcrypt hashes")
	}
}

// TestComparePassword tests password comparison
func TestComparePassword(t *testing.T) {
	validPassword := "SecureP@ss1"
	validHash, err := HashPassword(validPassword)
	if err != nil {
		t.Fatalf("Failed to create hash for testing: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password matches hash",
			hashedPassword: validHash,
			password:       validPassword,
			wantErr:        false,
		},
		{
			name:           "incorrect password fails",
			hashedPassword: validHash,
			password:       "WrongPassword123!",
			wantErr:        true,
		},
		{
			name:           "empty password fails",
			hashedPassword: validHash,
			password:       "",
			wantErr:        true,
		},
		{
			name:           "different password fails",
			hashedPassword: validHash,
			password:       "AnotherPass123!",
			wantErr:        true,
		},
		{
			name:           "case-sensitive comparison fails",
			hashedPassword: validHash,
			password:       "securep@ss1",
			wantErr:        true,
		},
		{
			name:           "invalid hash format returns error",
			hashedPassword: "invalid_hash",
			password:       validPassword,
			wantErr:        true,
		},
		{
			name:           "empty hash returns error",
			hashedPassword: "",
			password:       validPassword,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ComparePassword(tt.hashedPassword, tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ComparePassword() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("ComparePassword() unexpected error = %v", err)
				}
			}
		})
	}
}

// TestComparePassword_Unicode tests unicode password comparison
func TestComparePassword_Unicode(t *testing.T) {
	unicodePassword := "Pāss123!"
	hash, err := HashPassword(unicodePassword)
	if err != nil {
		t.Fatalf("Failed to create hash for unicode password: %v", err)
	}

	// Test that unicode password matches correctly
	err = ComparePassword(hash, unicodePassword)
	if err != nil {
		t.Errorf("ComparePassword() failed with unicode password: %v", err)
	}

	// Test that similar but different unicode password fails
	differentUnicodePassword := "Päss123!"
	err = ComparePassword(hash, differentUnicodePassword)
	if err == nil {
		t.Errorf("ComparePassword() should fail with different unicode password")
	}
}

// TestHasRepeatedChars tests the hasRepeatedChars helper function
func TestHasRepeatedChars(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "four same chars",
			password: "aaaa",
			want:     true,
		},
		{
			name:     "five same chars",
			password: "aaaaa",
			want:     true,
		},
		{
			name:     "three same chars (allowed)",
			password: "aaa",
			want:     false,
		},
		{
			name:     "two same chars (allowed)",
			password: "aa",
			want:     false,
		},
		{
			name:     "no repeated chars",
			password: "abcd",
			want:     false,
		},
		{
			name:     "short string (< 4 chars)",
			password: "abc",
			want:     false,
		},
		{
			name:     "repeated in middle",
			password: "abccccde",
			want:     true,
		},
		{
			name:     "different chars",
			password: "abcdef",
			want:     false,
		},
		{
			name:     "repeated at start",
			password: "aaaabc",
			want:     true,
		},
		{
			name:     "repeated at end",
			password: "abcdeeee",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasRepeatedChars(tt.password); got != tt.want {
				t.Errorf("hasRepeatedChars() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestPasswordValidationError tests the PasswordValidationError type
func TestPasswordValidationError(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{
			name:    "error message is correct",
			message: "Password must be at least 8 characters long",
		},
		{
			name:    "different error message",
			message: "Password must contain at least one uppercase letter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &PasswordValidationError{Message: tt.message}

			// Test that it implements error interface
			if err.Error() != tt.message {
				t.Errorf("PasswordValidationError.Error() = %v, want %v", err.Error(), tt.message)
			}

			// Test that it can be type-asserted
			if _, ok := interface{}(err).(*PasswordValidationError); !ok {
				t.Errorf("PasswordValidationError should be type-assertable")
			}
		})
	}
}

// TestIsPasswordValidationError tests the IsPasswordValidationError function
func TestIsPasswordValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "PasswordValidationError returns true",
			err:  &PasswordValidationError{Message: "test error"},
			want: true,
		},
		{
			name: "nil error returns false",
			err:  nil,
			want: false,
		},
		{
			name: "other error type returns false",
			err:  bcrypt.ErrMismatchedHashAndPassword,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPasswordValidationError(tt.err); got != tt.want {
				t.Errorf("IsPasswordValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIntegration_HashAndCompare tests the complete hash-compare workflow
func TestIntegration_HashAndCompare(t *testing.T) {
	password := "SecureP@ss1"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	// Compare with correct password
	err = ComparePassword(hash, password)
	if err != nil {
		t.Errorf("ComparePassword() with correct password failed: %v", err)
	}

	// Compare with wrong password
	wrongPassword := "WrongPassword123!"
	err = ComparePassword(hash, wrongPassword)
	if err == nil {
		t.Errorf("ComparePassword() with wrong password should fail")
	}

	// Test that hash is not reversible
	if hash == password {
		t.Errorf("Hash should not equal original password")
	}
}

// TestIntegration_MultipleUsers tests that multiple users with same password have different hashes
func TestIntegration_MultipleUsers(t *testing.T) {
	password := "SecureUser123!" // Changed from "CommonPassword123!"

	// Hash the same password multiple times (simulating different users)
	hash1, err1 := HashPassword(password)
	if err1 != nil {
		t.Fatalf("HashPassword() first attempt failed: %v", err1)
	}

	hash2, err2 := HashPassword(password)
	if err2 != nil {
		t.Fatalf("HashPassword() second attempt failed: %v", err2)
	}

	hash3, err3 := HashPassword(password)
	if err3 != nil {
		t.Fatalf("HashPassword() third attempt failed: %v", err3)
	}

	// All hashes should be different
	if hash1 == hash2 || hash1 == hash3 || hash2 == hash3 {
		t.Errorf("Multiple users with same password should have different hashes")
	}

	// All hashes should work with the original password
	hashes := []string{hash1, hash2, hash3}
	for i, hash := range hashes {
		err := ComparePassword(hash, password)
		if err != nil {
			t.Errorf("ComparePassword() failed for hash %d: %v", i+1, err)
		}
	}
}

// BenchmarkHashPassword benchmarks password hashing performance
func BenchmarkHashPassword(b *testing.B) {
	password := "SecureP@ss1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword() failed: %v", err)
		}
	}
}

// BenchmarkValidatePassword benchmarks password validation performance
func BenchmarkValidatePassword(b *testing.B) {
	password := "SecureP@ss1"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ValidatePassword(password)
		if err != nil {
			b.Fatalf("ValidatePassword() failed: %v", err)
		}
	}
}

// BenchmarkComparePassword benchmarks password comparison performance
func BenchmarkComparePassword(b *testing.B) {
	password := "SecureP@ss1"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to create hash for benchmarking: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := ComparePassword(hash, password)
		if err != nil {
			b.Fatalf("ComparePassword() failed: %v", err)
		}
	}
}

// BenchmarkHasRepeatedChars benchmarks repeated character detection
func BenchmarkHasRepeatedChars(b *testing.B) {
	password := "Pass1111!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasRepeatedChars(password)
	}
}
