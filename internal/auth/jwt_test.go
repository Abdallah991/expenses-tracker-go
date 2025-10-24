package auth

import (
	"os"
	"testing"
	"time"
)

func setupJWTTest() {
	// Set up test JWT configuration
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("JWT_ACCESS_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_EXPIRY", "7d")
	InitJWT()
}

func TestInitJWT_Success(t *testing.T) {
	// Set environment variables
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "30m")
	os.Setenv("JWT_REFRESH_EXPIRY", "336h") // 14 days in hours

	err := InitJWT()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify configuration was set
	if jwtSecret != "test-secret" {
		t.Errorf("Expected jwtSecret to be 'test-secret', got %s", jwtSecret)
	}

	if accessExpiry != 30*time.Minute {
		t.Errorf("Expected accessExpiry to be 30m, got %v", accessExpiry)
	}

	if refreshExpiry != 14*24*time.Hour {
		t.Errorf("Expected refreshExpiry to be 14d, got %v", refreshExpiry)
	}
}

func TestInitJWT_DefaultValues(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_ACCESS_EXPIRY")
	os.Unsetenv("JWT_REFRESH_EXPIRY")

	// Set only required JWT_SECRET
	os.Setenv("JWT_SECRET", "test-secret")

	err := InitJWT()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify default values
	if accessExpiry != 15*time.Minute {
		t.Errorf("Expected default accessExpiry to be 15m, got %v", accessExpiry)
	}

	if refreshExpiry != 7*24*time.Hour {
		t.Errorf("Expected default refreshExpiry to be 7d, got %v", refreshExpiry)
	}
}

func TestInitJWT_MissingSecret(t *testing.T) {
	// Clear JWT_SECRET
	os.Unsetenv("JWT_SECRET")

	err := InitJWT()
	if err == nil {
		t.Error("Expected error for missing JWT_SECRET, got nil")
	}

	if err.Error() != "JWT_SECRET environment variable is required" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestInitJWT_InvalidAccessExpiry(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "invalid-duration")

	err := InitJWT()
	if err == nil {
		t.Error("Expected error for invalid JWT_ACCESS_EXPIRY, got nil")
	}

	if err.Error()[:len("invalid JWT_ACCESS_EXPIRY format")] != "invalid JWT_ACCESS_EXPIRY format" {
		t.Errorf("Expected error about invalid format, got %s", err.Error())
	}
}

func TestInitJWT_InvalidRefreshExpiry(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_REFRESH_EXPIRY", "invalid-duration")

	err := InitJWT()
	if err == nil {
		t.Error("Expected error for invalid JWT_REFRESH_EXPIRY, got nil")
	}

	// The error message should contain information about invalid duration
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestGenerateAccessToken_Success(t *testing.T) {
	setupJWTTest()

	token, err := GenerateAccessToken(123, "test@example.com")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token, got empty string")
	}

	// Token should be a valid JWT format (3 parts separated by dots)
	parts := len([]rune(token))
	if parts < 100 { // JWT tokens are typically much longer
		t.Errorf("Token seems too short: %d characters", parts)
	}
}

func TestGenerateAccessToken_DifferentUsers(t *testing.T) {
	setupJWTTest()

	// Generate tokens for different users
	token1, err1 := GenerateAccessToken(123, "user1@example.com")
	token2, err2 := GenerateAccessToken(456, "user2@example.com")

	if err1 != nil {
		t.Errorf("Expected no error for user1, got %v", err1)
	}
	if err2 != nil {
		t.Errorf("Expected no error for user2, got %v", err2)
	}

	// Tokens should be different
	if token1 == token2 {
		t.Error("Expected different tokens for different users")
	}
}

func TestValidateAccessToken_Success(t *testing.T) {
	setupJWTTest()

	// Generate a token
	originalToken, err := GenerateAccessToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate the token
	claims, err := ValidateAccessToken(originalToken)
	if err != nil {
		t.Errorf("Expected no error validating token, got %v", err)
	}

	if claims.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", claims.Email)
	}
}

func TestValidateAccessToken_InvalidToken(t *testing.T) {
	setupJWTTest()

	// Test with invalid token
	_, err := ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateAccessToken_EmptyToken(t *testing.T) {
	setupJWTTest()

	// Test with empty token
	_, err := ValidateAccessToken("")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	setupJWTTest()

	// Generate token with current secret
	token, err := GenerateAccessToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Change secret
	os.Setenv("JWT_SECRET", "different-secret")
	InitJWT()

	// Try to validate with different secret
	_, err = ValidateAccessToken(token)
	if err == nil {
		t.Error("Expected error for token with wrong secret, got nil")
	}
}

func TestExtractTokenFromHeader_Success(t *testing.T) {
	token, err := ExtractTokenFromHeader("Bearer valid-token-here")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if token != "valid-token-here" {
		t.Errorf("Expected token 'valid-token-here', got %s", token)
	}
}

func TestExtractTokenFromHeader_EmptyHeader(t *testing.T) {
	_, err := ExtractTokenFromHeader("")
	if err == nil {
		t.Error("Expected error for empty header, got nil")
	}

	if err.Error() != "authorization header is required" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestExtractTokenFromHeader_WrongPrefix(t *testing.T) {
	_, err := ExtractTokenFromHeader("Basic token-here")
	if err == nil {
		t.Error("Expected error for wrong prefix, got nil")
	}

	if err.Error() != "authorization header must start with 'Bearer '" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestExtractTokenFromHeader_NoPrefix(t *testing.T) {
	_, err := ExtractTokenFromHeader("token-here")
	if err == nil {
		t.Error("Expected error for no prefix, got nil")
	}

	if err.Error() != "authorization header must start with 'Bearer '" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestExtractTokenFromHeader_EmptyToken(t *testing.T) {
	_, err := ExtractTokenFromHeader("Bearer ")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}

	if err.Error() != "token cannot be empty" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestExtractTokenFromHeader_ShortHeader(t *testing.T) {
	_, err := ExtractTokenFromHeader("Bear")
	if err == nil {
		t.Error("Expected error for short header, got nil")
	}

	if err.Error() != "authorization header must start with 'Bearer '" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestGenerateRefreshToken_Success(t *testing.T) {
	token, err := GenerateRefreshToken()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty refresh token, got empty string")
	}

	// Refresh token should be 64 characters (32 bytes hex encoded)
	if len(token) != 64 {
		t.Errorf("Expected refresh token length 64, got %d", len(token))
	}
}

func TestGenerateRefreshToken_Uniqueness(t *testing.T) {
	// Generate multiple refresh tokens
	token1, err1 := GenerateRefreshToken()
	token2, err2 := GenerateRefreshToken()

	if err1 != nil {
		t.Errorf("Expected no error for token1, got %v", err1)
	}
	if err2 != nil {
		t.Errorf("Expected no error for token2, got %v", err2)
	}

	// Tokens should be different
	if token1 == token2 {
		t.Error("Expected different refresh tokens")
	}
}

func TestGenerateSecureToken_Success(t *testing.T) {
	token, err := GenerateSecureToken(16)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token, got empty string")
	}

	// 16 bytes should produce 32 hex characters
	if len(token) != 32 {
		t.Errorf("Expected token length 32, got %d", len(token))
	}
}

func TestGenerateSecureToken_DifferentLengths(t *testing.T) {
	testCases := []struct {
		length      int
		expectedLen int
	}{
		{8, 16},   // 8 bytes = 16 hex chars
		{16, 32},  // 16 bytes = 32 hex chars
		{32, 64},  // 32 bytes = 64 hex chars
		{64, 128}, // 64 bytes = 128 hex chars
	}

	for _, tc := range testCases {
		t.Run("length_"+string(rune(tc.length)), func(t *testing.T) {
			token, err := GenerateSecureToken(tc.length)
			if err != nil {
				t.Errorf("Expected no error for length %d, got %v", tc.length, err)
			}

			if len(token) != tc.expectedLen {
				t.Errorf("Expected length %d for input %d, got %d", tc.expectedLen, tc.length, len(token))
			}
		})
	}
}

func TestGenerateSecureToken_ZeroLength(t *testing.T) {
	token, err := GenerateSecureToken(0)
	if err != nil {
		t.Errorf("Expected no error for zero length, got %v", err)
	}

	if token != "" {
		t.Errorf("Expected empty token for zero length, got %s", token)
	}
}

func TestGetTokenExpiry(t *testing.T) {
	setupJWTTest()

	expiry := GetTokenExpiry()
	if expiry != 15*time.Minute {
		t.Errorf("Expected expiry 15m, got %v", expiry)
	}
}

func TestGetRefreshTokenExpiry(t *testing.T) {
	// Set up test environment
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_REFRESH_EXPIRY", "168h") // 7 days in hours
	InitJWT()

	expiry := GetRefreshTokenExpiry()
	if expiry != 7*24*time.Hour {
		t.Errorf("Expected expiry 7d, got %v", expiry)
	}
}

func TestJWTClaims_Structure(t *testing.T) {
	claims := JWTClaims{
		UserID: 123,
		Email:  "test@example.com",
	}

	if claims.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", claims.Email)
	}
}

func TestTokenPair_Structure(t *testing.T) {
	tokenPair := TokenPair{
		AccessToken:  "access-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    3600,
	}

	if tokenPair.AccessToken != "access-token" {
		t.Errorf("Expected AccessToken 'access-token', got %s", tokenPair.AccessToken)
	}

	if tokenPair.RefreshToken != "refresh-token" {
		t.Errorf("Expected RefreshToken 'refresh-token', got %s", tokenPair.RefreshToken)
	}

	if tokenPair.ExpiresIn != 3600 {
		t.Errorf("Expected ExpiresIn 3600, got %d", tokenPair.ExpiresIn)
	}
}

// Integration test: Generate and validate token
func TestIntegration_GenerateAndValidate(t *testing.T) {
	setupJWTTest()

	// Generate token
	originalToken, err := GenerateAccessToken(789, "integration@test.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token
	claims, err := ValidateAccessToken(originalToken)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Verify claims
	if claims.UserID != 789 {
		t.Errorf("Expected UserID 789, got %d", claims.UserID)
	}

	if claims.Email != "integration@test.com" {
		t.Errorf("Expected email 'integration@test.com', got %s", claims.Email)
	}

	// Verify standard claims
	if claims.Issuer != "expenses-tracker-go" {
		t.Errorf("Expected issuer 'expenses-tracker-go', got %s", claims.Issuer)
	}

	if claims.Subject != "789" {
		t.Errorf("Expected subject '789', got %s", claims.Subject)
	}
}

// Test token expiration
func TestTokenExpiration(t *testing.T) {
	// Set very short expiry
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("JWT_ACCESS_EXPIRY", "1ms")
	InitJWT()

	// Generate token
	token, err := GenerateAccessToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Try to validate expired token
	_, err = ValidateAccessToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

// Test token not yet valid (NotBefore)
func TestTokenNotBefore(t *testing.T) {
	setupJWTTest()

	// Generate token
	token, err := GenerateAccessToken(123, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Token should be valid immediately
	_, err = ValidateAccessToken(token)
	if err != nil {
		t.Errorf("Expected token to be valid, got error: %v", err)
	}
}

// Test multiple token generation uniqueness
func TestTokenUniqueness(t *testing.T) {
	setupJWTTest()

	tokens := make(map[string]bool)

	// Generate 100 tokens
	for i := 0; i < 100; i++ {
		token, err := GenerateAccessToken(i, "test@example.com")
		if err != nil {
			t.Fatalf("Failed to generate token %d: %v", i, err)
		}

		// Check for uniqueness
		if tokens[token] {
			t.Errorf("Duplicate token generated at iteration %d", i)
		}
		tokens[token] = true
	}
}

// Test refresh token uniqueness
func TestRefreshTokenUniqueness(t *testing.T) {
	tokens := make(map[string]bool)

	// Generate 100 refresh tokens
	for i := 0; i < 100; i++ {
		token, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("Failed to generate refresh token %d: %v", i, err)
		}

		// Check for uniqueness
		if tokens[token] {
			t.Errorf("Duplicate refresh token generated at iteration %d", i)
		}
		tokens[token] = true
	}
}
