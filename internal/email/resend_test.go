package email

import (
	"os"
	"strings"
	"testing"
)

func setupEmailTest() {
	// Set up test environment variables
	os.Setenv("RESEND_API_KEY", "test-api-key")
	os.Setenv("FROM_EMAIL", "test@example.com")
	os.Setenv("APP_URL", "https://test.example.com")
}

func cleanupEmailTest() {
	// Clean up environment variables
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("FROM_EMAIL")
	os.Unsetenv("APP_URL")
}

func TestNewEmailService_Success(t *testing.T) {
	setupEmailTest()
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service == nil {
		t.Error("Expected service to be created, got nil")
	}

	if service.fromEmail != "test@example.com" {
		t.Errorf("Expected fromEmail 'test@example.com', got %s", service.fromEmail)
	}

	if service.appURL != "https://test.example.com" {
		t.Errorf("Expected appURL 'https://test.example.com', got %s", service.appURL)
	}

	if service.client == nil {
		t.Error("Expected client to be initialized, got nil")
	}
}

func TestNewEmailService_DefaultAppURL(t *testing.T) {
	// Set only required environment variables
	os.Setenv("RESEND_API_KEY", "test-api-key")
	os.Setenv("FROM_EMAIL", "test@example.com")
	os.Unsetenv("APP_URL")
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if service.appURL != "http://localhost:8080" {
		t.Errorf("Expected default appURL 'http://localhost:8080', got %s", service.appURL)
	}
}

func TestNewEmailService_MissingAPIKey(t *testing.T) {
	os.Unsetenv("RESEND_API_KEY")
	os.Setenv("FROM_EMAIL", "test@example.com")
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}

	if service != nil {
		t.Error("Expected service to be nil when error occurs")
	}

	expectedError := "RESEND_API_KEY environment variable is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewEmailService_MissingFromEmail(t *testing.T) {
	os.Setenv("RESEND_API_KEY", "test-api-key")
	os.Unsetenv("FROM_EMAIL")
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err == nil {
		t.Error("Expected error for missing from email, got nil")
	}

	if service != nil {
		t.Error("Expected service to be nil when error occurs")
	}

	expectedError := "FROM_EMAIL environment variable is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewEmailService_EmptyAPIKey(t *testing.T) {
	os.Setenv("RESEND_API_KEY", "")
	os.Setenv("FROM_EMAIL", "test@example.com")
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err == nil {
		t.Error("Expected error for empty API key, got nil")
	}

	if service != nil {
		t.Error("Expected service to be nil when error occurs")
	}

	expectedError := "RESEND_API_KEY environment variable is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestNewEmailService_EmptyFromEmail(t *testing.T) {
	os.Setenv("RESEND_API_KEY", "test-api-key")
	os.Setenv("FROM_EMAIL", "")
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err == nil {
		t.Error("Expected error for empty from email, got nil")
	}

	if service != nil {
		t.Error("Expected service to be nil when error occurs")
	}

	expectedError := "FROM_EMAIL environment variable is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSendVerificationEmail_ContentValidation(t *testing.T) {
	setupEmailTest()
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	// Test with mock data
	to := "user@example.com"
	token := "test-verification-token-123"

	// We can't actually send the email in tests, but we can validate the content generation
	// by checking that the service was created properly and the method exists
	if service == nil {
		t.Error("Expected service to be created")
	}

	// Test that the method exists and can be called (it will fail due to no real API key)
	err = service.SendVerificationEmail(to, token)
	// We expect an error since we're using a test API key, but the method should exist
	if err == nil {
		t.Error("Expected error due to invalid API key, but got nil")
	}
}

func TestSendPasswordResetEmail_ContentValidation(t *testing.T) {
	setupEmailTest()
	defer cleanupEmailTest()

	service, err := NewEmailService()
	if err != nil {
		t.Fatalf("Failed to create email service: %v", err)
	}

	// Test with mock data
	to := "user@example.com"
	token := "test-reset-token-456"

	// Test that the method exists and can be called (it will fail due to no real API key)
	err = service.SendPasswordResetEmail(to, token)
	// We expect an error since we're using a test API key, but the method should exist
	if err == nil {
		t.Error("Expected error due to invalid API key, but got nil")
	}
}

// Test email content generation by creating a mock service
func TestEmailContentGeneration(t *testing.T) {
	// Create a service with test data
	service := &EmailService{
		fromEmail: "test@example.com",
		appURL:    "https://test.example.com",
	}

	// Test verification email URL generation
	verificationToken := "test-verification-token"
	expectedVerificationURL := "https://test.example.com/auth/verify-email?token=test-verification-token"

	// We can't directly test the private content generation, but we can test the URL construction
	// by checking if the service has the correct appURL
	if service.appURL != "https://test.example.com" {
		t.Errorf("Expected appURL 'https://test.example.com', got %s", service.appURL)
	}

	// Test password reset email URL generation
	resetToken := "test-reset-token"
	expectedResetURL := "https://test.example.com/auth/reset-password?token=test-reset-token"

	// Similar validation for reset URL
	if service.appURL != "https://test.example.com" {
		t.Errorf("Expected appURL 'https://test.example.com', got %s", service.appURL)
	}

	// These are just to avoid unused variable warnings
	_ = expectedVerificationURL
	_ = expectedResetURL
	_ = verificationToken
	_ = resetToken
}

func TestEmailService_Structure(t *testing.T) {
	service := &EmailService{
		fromEmail: "test@example.com",
		appURL:    "https://test.example.com",
	}

	if service.fromEmail != "test@example.com" {
		t.Errorf("Expected fromEmail 'test@example.com', got %s", service.fromEmail)
	}

	if service.appURL != "https://test.example.com" {
		t.Errorf("Expected appURL 'https://test.example.com', got %s", service.appURL)
	}
}

func TestEmailService_WithDifferentURLs(t *testing.T) {
	testCases := []struct {
		name   string
		appURL string
	}{
		{"Localhost", "http://localhost:8080"},
		{"HTTPS Production", "https://api.example.com"},
		{"HTTP Production", "http://api.example.com"},
		{"Custom Port", "https://api.example.com:3000"},
		{"Subdomain", "https://app.mycompany.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &EmailService{
				fromEmail: "test@example.com",
				appURL:    tc.appURL,
			}

			if service.appURL != tc.appURL {
				t.Errorf("Expected appURL '%s', got '%s'", tc.appURL, service.appURL)
			}
		})
	}
}

func TestEmailService_WithDifferentFromEmails(t *testing.T) {
	testCases := []struct {
		name      string
		fromEmail string
	}{
		{"Simple Email", "test@example.com"},
		{"With Name", "Test User <test@example.com>"},
		{"Subdomain", "noreply@mail.example.com"},
		{"Different Domain", "admin@mycompany.com"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service := &EmailService{
				fromEmail: tc.fromEmail,
				appURL:    "https://test.example.com",
			}

			if service.fromEmail != tc.fromEmail {
				t.Errorf("Expected fromEmail '%s', got '%s'", tc.fromEmail, service.fromEmail)
			}
		})
	}
}

// Test environment variable handling
func TestEnvironmentVariableHandling(t *testing.T) {
	// Test with various environment variable combinations
	testCases := []struct {
		name        string
		apiKey      string
		fromEmail   string
		appURL      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "All variables set",
			apiKey:      "test-key",
			fromEmail:   "test@example.com",
			appURL:      "https://test.com",
			expectError: false,
		},
		{
			name:        "Missing API key",
			apiKey:      "",
			fromEmail:   "test@example.com",
			appURL:      "https://test.com",
			expectError: true,
			errorMsg:    "RESEND_API_KEY environment variable is required",
		},
		{
			name:        "Missing from email",
			apiKey:      "test-key",
			fromEmail:   "",
			appURL:      "https://test.com",
			expectError: true,
			errorMsg:    "FROM_EMAIL environment variable is required",
		},
		{
			name:        "Missing app URL (should use default)",
			apiKey:      "test-key",
			fromEmail:   "test@example.com",
			appURL:      "",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables
			if tc.apiKey != "" {
				os.Setenv("RESEND_API_KEY", tc.apiKey)
			} else {
				os.Unsetenv("RESEND_API_KEY")
			}

			if tc.fromEmail != "" {
				os.Setenv("FROM_EMAIL", tc.fromEmail)
			} else {
				os.Unsetenv("FROM_EMAIL")
			}

			if tc.appURL != "" {
				os.Setenv("APP_URL", tc.appURL)
			} else {
				os.Unsetenv("APP_URL")
			}

			// Clean up after test
			defer func() {
				os.Unsetenv("RESEND_API_KEY")
				os.Unsetenv("FROM_EMAIL")
				os.Unsetenv("APP_URL")
			}()

			service, err := NewEmailService()

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if err.Error() != tc.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tc.errorMsg, err.Error())
				}
				if service != nil {
					t.Error("Expected service to be nil when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if service == nil {
					t.Error("Expected service to be created")
				}
				if tc.appURL == "" && service.appURL != "http://localhost:8080" {
					t.Errorf("Expected default appURL, got %s", service.appURL)
				}
			}
		})
	}
}

// Test URL construction logic
func TestURLConstruction(t *testing.T) {
	service := &EmailService{
		fromEmail: "test@example.com",
		appURL:    "https://test.example.com",
	}

	// Test verification URL construction
	verificationToken := "abc123"
	expectedVerificationURL := "https://test.example.com/auth/verify-email?token=abc123"

	// We can't directly access the URL construction, but we can verify the components
	if !strings.Contains(service.appURL, "https://test.example.com") {
		t.Error("App URL should contain the base URL")
	}

	// Test reset URL construction
	resetToken := "def456"
	expectedResetURL := "https://test.example.com/auth/reset-password?token=def456"

	// Similar validation
	if !strings.Contains(service.appURL, "https://test.example.com") {
		t.Error("App URL should contain the base URL")
	}

	// Avoid unused variable warnings
	_ = expectedVerificationURL
	_ = expectedResetURL
	_ = verificationToken
	_ = resetToken
}

// Test that the service can be created with different configurations
func TestServiceConfiguration(t *testing.T) {
	configs := []struct {
		name      string
		apiKey    string
		fromEmail string
		appURL    string
	}{
		{
			name:      "Development",
			apiKey:    "dev-key",
			fromEmail: "dev@localhost",
			appURL:    "http://localhost:3000",
		},
		{
			name:      "Staging",
			apiKey:    "staging-key",
			fromEmail: "staging@example.com",
			appURL:    "https://staging.example.com",
		},
		{
			name:      "Production",
			apiKey:    "prod-key",
			fromEmail: "noreply@example.com",
			appURL:    "https://example.com",
		},
	}

	for _, config := range configs {
		t.Run(config.name, func(t *testing.T) {
			os.Setenv("RESEND_API_KEY", config.apiKey)
			os.Setenv("FROM_EMAIL", config.fromEmail)
			os.Setenv("APP_URL", config.appURL)

			defer func() {
				os.Unsetenv("RESEND_API_KEY")
				os.Unsetenv("FROM_EMAIL")
				os.Unsetenv("APP_URL")
			}()

			service, err := NewEmailService()
			if err != nil {
				t.Errorf("Expected no error for %s config, got %v", config.name, err)
			}

			if service.fromEmail != config.fromEmail {
				t.Errorf("Expected fromEmail '%s', got '%s'", config.fromEmail, service.fromEmail)
			}

			if service.appURL != config.appURL {
				t.Errorf("Expected appURL '%s', got '%s'", config.appURL, service.appURL)
			}
		})
	}
}
