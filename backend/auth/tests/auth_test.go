package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Assume you have a RegisterHandler function in your main code
// import your handler package if needed

func TestRegisterRoute(t *testing.T) {
	router := setupRouter()

	// Prepare input JSON
	input := struct {
		Email           string
		Password        string
		ConfirmPassword string
	}{
		Email:           "test@example.com",
		Password:        "securepassword",
		ConfirmPassword: "securepassword",
	}
	body, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("Failed to marshal input: %v", err)
	}

	// Create a request to the register route
	req, err := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Use httptest to record the response
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	// Check status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 OK, got %d", rr.Code)
	}

	// Parse response JSON
	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check for JWT token in response
	if _, ok := resp["token"]; !ok {
		t.Errorf("Expected 'token' in response, got %v", resp)
	}
}
