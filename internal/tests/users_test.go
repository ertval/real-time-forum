package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"forum/internal/router"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUserRegistration(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	r := router.NewRouter(db)
	// Test valid registration
	validBody := `{"username":"newuser","email":"new@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBufferString(validBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	id, ok := resp["data"].(map[string]interface{})["id"]
	if !ok {
		t.Errorf("expected id in response")
	}
	// Login the user
	loginBody := `{"username":"newuser","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}
	// Extract token
	setCookie := rec.Header().Get("Set-Cookie")
	parts := strings.Split(setCookie, ";")
	token := strings.TrimPrefix(strings.TrimSpace(parts[0]), "session_token=")

	// Test user retrieval
	userID := int64(id.(float64))
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200 for user get, got %d", rec.Code)
	}
	var userResp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&userResp); err != nil {
		t.Fatalf("failed to decode user response: %v", err)
	}
	data, ok := userResp["data"].(map[string]interface{})
	if !ok || data["username"] != "newuser" {
		t.Errorf("expected user data with username 'newuser'")
	}
	// Test invalid cases
	invalidCases := []struct {
		body string
		desc string
	}{
		{`{"username":"ab","email":"test@example.com","password":"password123"}`, "short username"},
		{`{"username":"newuser2","email":"test@example.com","password":"123"}`, "weak password"},
		{`{"username":"newuser2","email":"invalid","password":"password123"}`, "invalid email"},
		{`{"username":"testuser","email":"dup@example.com","password":"password123"}`, "duplicate username"},
	}
	for _, tc := range invalidCases {
		req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for %s, got %d", tc.desc, rec.Code)
		}
	}
}

func TestAuthFlow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	r := router.NewRouter(db)
	// Register a user
	regBody := `{"username":"newuser123","email":"test2@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	t.Logf("register response body: %s", rec.Body.String())
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 for register, got %d", rec.Code)
	}
	// Login
	loginBody := `{"username":"newuser123","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	t.Logf("login response body: %s", rec.Body.String())
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for login, got %d", rec.Code)
	}
	// Check login response has success message
	var loginResp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	userData, ok := loginResp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in login response")
	}
	message, ok := userData["message"].(string)
	if !ok || message != "Login successful" {
		t.Fatalf("expected 'Login successful' message, got %v", userData)
	}
	// Extract token from Set-Cookie header
	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie header")
	}
	// Parse cookie: "session_token=uuid; Path=/; HttpOnly; SameSite=StrictMode"
	parts := strings.Split(setCookie, ";")
	tokenPart := strings.TrimSpace(parts[0])
	if !strings.HasPrefix(tokenPart, "session_token=") {
		t.Fatalf("expected session_token in cookie")
	}
	token := strings.TrimPrefix(tokenPart, "session_token=")
	// Call /me with cookie
	req = httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 for /me, got %d", rec.Code)
	}
	var meResp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&meResp); err != nil {
		t.Fatalf("failed to decode /me response: %v", err)
	}
	data, ok := meResp["data"].(map[string]interface{})
	if !ok || data["username"] != "newuser123" {
		t.Errorf("expected user data with username 'newuser123'")
	}
}

func TestUserLogout(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register + Login
	regBody := `{"username":"logoutUser","email":"logout@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"logoutUser","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected cookie on login")
	}
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// Logout
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/logout", nil)
	req.Header.Set("Cookie", "session_token="+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 logout, got %d", w.Code)
	}

	// Try calling /me → should fail
	req2 := httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req2.Header.Set("Cookie", "session_token="+token)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after logout, got %d", w2.Code)
	}
}
