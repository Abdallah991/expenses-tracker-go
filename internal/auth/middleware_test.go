package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

//! this doesnt include integration test with database

func setupTest() {
	// Set up test JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only")
	os.Setenv("JWT_ACCESS_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_EXPIRY", "7d")
	InitJWT()
}

// Test handler that checks if user context is set
func testHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "User ID not found", http.StatusInternalServerError)
		return
	}

	email, err := GetUserEmailFromContext(r.Context())
	if err != nil {
		http.Error(w, "User email not found", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"user_id": userID,
		"email":   email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Test handler that works without authentication for OptionalAuth tests
func testHandlerOptional(w http.ResponseWriter, r *http.Request) {
	userID, err := GetUserIDFromContext(r.Context())
	if err != nil {
		// No user context - this is expected for OptionalAuth
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "anonymous"})
		return
	}

	email, _ := GetUserEmailFromContext(r.Context())
	response := map[string]interface{}{
		"status":  "authenticated",
		"user_id": userID,
		"email":   email,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Test the context helper functions
func TestGetUserIDFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)

	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if userID != 123 {
		t.Errorf("Expected user ID 123, got %d", userID)
	}
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := GetUserIDFromContext(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "user ID not found in context" {
		t.Errorf("Expected 'user ID not found in context', got %s", err.Error())
	}
}

func TestGetUserIDFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "not-an-int")

	_, err := GetUserIDFromContext(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "user ID not found in context" {
		t.Errorf("Expected 'user ID not found in context', got %s", err.Error())
	}
}

func TestGetUserEmailFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserEmailKey, "test@example.com")

	email, err := GetUserEmailFromContext(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", email)
	}
}

func TestGetUserEmailFromContext_NotFound(t *testing.T) {
	ctx := context.Background()

	_, err := GetUserEmailFromContext(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "user email not found in context" {
		t.Errorf("Expected 'user email not found in context', got %s", err.Error())
	}
}

func TestGetUserEmailFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserEmailKey, 123)

	_, err := GetUserEmailFromContext(ctx)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "user email not found in context" {
		t.Errorf("Expected 'user email not found in context', got %s", err.Error())
	}
}

func TestWriteErrorResponse(t *testing.T) {
	rr := httptest.NewRecorder()

	writeErrorResponse(rr, http.StatusBadRequest, "Test error", "Test details")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got %s", contentType)
	}

	var errorResp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Error != "Test error" {
		t.Errorf("Expected error 'Test error', got %s", errorResp.Error)
	}

	if errorResp.Details != "Test details" {
		t.Errorf("Expected details 'Test details', got %s", errorResp.Details)
	}
}

func TestWriteErrorResponse_NoDetails(t *testing.T) {
	rr := httptest.NewRecorder()

	writeErrorResponse(rr, http.StatusBadRequest, "Test error", "")

	var errorResp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResp.Details != "" {
		t.Errorf("Expected empty details, got %s", errorResp.Details)
	}
}

func TestWriteErrorResponse_DifferentStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode int
		error      string
		details    string
	}{
		{http.StatusUnauthorized, "Unauthorized", "Invalid token"},
		{http.StatusForbidden, "Forbidden", "Access denied"},
		{http.StatusInternalServerError, "Internal Server Error", "Database error"},
		{http.StatusNotFound, "Not Found", "Resource not found"},
	}

	for _, tc := range testCases {
		t.Run(tc.error, func(t *testing.T) {
			rr := httptest.NewRecorder()

			writeErrorResponse(rr, tc.statusCode, tc.error, tc.details)

			if rr.Code != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, rr.Code)
			}

			var errorResp ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &errorResp); err != nil {
				t.Fatalf("Failed to unmarshal error response: %v", err)
			}

			if errorResp.Error != tc.error {
				t.Errorf("Expected error '%s', got '%s'", tc.error, errorResp.Error)
			}

			if errorResp.Details != tc.details {
				t.Errorf("Expected details '%s', got '%s'", tc.details, errorResp.Details)
			}
		})
	}
}

// Test context key types to ensure they work correctly
func TestContextKeyTypes(t *testing.T) {
	// Test that our context keys are the correct type
	if string(UserIDKey) != "user_id" {
		t.Errorf("Expected UserIDKey to be 'user_id', got %s", string(UserIDKey))
	}

	if string(UserEmailKey) != "user_email" {
		t.Errorf("Expected UserEmailKey to be 'user_email', got %s", string(UserEmailKey))
	}
}

// Test ErrorResponse struct
func TestErrorResponse(t *testing.T) {
	errorResp := ErrorResponse{
		Error:   "Test error",
		Details: "Test details",
	}

	if errorResp.Error != "Test error" {
		t.Errorf("Expected error 'Test error', got %s", errorResp.Error)
	}

	if errorResp.Details != "Test details" {
		t.Errorf("Expected details 'Test details', got %s", errorResp.Details)
	}
}

// Test ErrorResponse JSON marshaling
func TestErrorResponse_JSONMarshaling(t *testing.T) {
	errorResp := ErrorResponse{
		Error:   "Test error",
		Details: "Test details",
	}

	jsonData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if unmarshaled.Error != errorResp.Error {
		t.Errorf("Expected error '%s', got '%s'", errorResp.Error, unmarshaled.Error)
	}

	if unmarshaled.Details != errorResp.Details {
		t.Errorf("Expected details '%s', got '%s'", errorResp.Details, unmarshaled.Details)
	}
}

// Test ErrorResponse with empty details
func TestErrorResponse_EmptyDetails(t *testing.T) {
	errorResp := ErrorResponse{
		Error:   "Test error",
		Details: "",
	}

	jsonData, err := json.Marshal(errorResp)
	if err != nil {
		t.Fatalf("Failed to marshal ErrorResponse: %v", err)
	}

	var unmarshaled ErrorResponse
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ErrorResponse: %v", err)
	}

	if unmarshaled.Details != "" {
		t.Errorf("Expected empty details, got '%s'", unmarshaled.Details)
	}
}

// Test context value setting and retrieval
func TestContextValueFlow(t *testing.T) {
	// Test setting multiple values in context
	ctx := context.Background()
	ctx = context.WithValue(ctx, UserIDKey, 123)
	ctx = context.WithValue(ctx, UserEmailKey, "test@example.com")

	// Test retrieving values
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		t.Errorf("Expected no error getting user ID, got %v", err)
	}

	if userID != 123 {
		t.Errorf("Expected user ID 123, got %d", userID)
	}

	email, err := GetUserEmailFromContext(ctx)
	if err != nil {
		t.Errorf("Expected no error getting user email, got %v", err)
	}

	if email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", email)
	}
}

// Test context key collision prevention
func TestContextKeyCollisionPrevention(t *testing.T) {
	// Test that different context key types don't interfere
	ctx := context.Background()

	// Set values with different key types
	ctx = context.WithValue(ctx, UserIDKey, 123)
	ctx = context.WithValue(ctx, "user_id", "value-for-testing")

	// Retrieve with our typed key
	userID, err := GetUserIDFromContext(ctx)
	if err != nil {
		t.Errorf("Expected no error getting user ID, got %v", err)
	}

	if userID != 123 {
		t.Errorf("Expected user ID 123, got %d", userID)
	}

	// Retrieve with string key
	stringValue := ctx.Value("user_id")
	if stringValue != "string-value" {
		t.Errorf("Expected string value 'string-value', got %v", stringValue)
	}
}
