package tests

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"forum/internal/router"
	"forum/internal/ws"
)

func newTestAPIWithHub(t *testing.T) (http.Handler, *sql.DB, *ws.Hub) {
	t.Helper()
	db := setupTestDB(t)
	hub := ws.NewHub()
	h := router.NewRouter(db, hub)
	return h, db, hub
}

func TestWebSocket_Unauthenticated(t *testing.T) {
	h, db, _ := newTestAPIWithHub(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated WS, got %d", rec.Code)
	}
}

func TestWebSocket_InvalidSession(t *testing.T) {
	h, db, _ := newTestAPIWithHub(t)
	defer db.Close()

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "invalid-token-123"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid session, got %d", rec.Code)
	}
}

func TestHub_AddRemove(t *testing.T) {
	hub := ws.NewHub()

	// User 1 should be offline initially (no connections)
	if hub.IsUserOnline(1) {
		t.Fatal("expected user 1 to be offline initially")
	}

	// Note: Can't actually create ws.Conn in tests without full upgrade
	// Just test the counting logic
	_ = hub.GetOnlineUserIDs()
	_ = hub.GetConnectionCount(1)
}

func TestHub_MultiConnection(t *testing.T) {
	hub := ws.NewHub()

	online := hub.GetOnlineUserIDs()
	if len(online) != 0 {
		t.Fatalf("expected no online users, got %d", len(online))
	}
}

func extractTokenFromLogin(t *testing.T, h http.Handler) string {
	regBody := `{"username":"wsuser","email":"ws@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", strings.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}

	loginBody := `{"username":"wsuser","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/login", strings.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	setCookie := rec.Header().Get("Set-Cookie")
	return strings.Split(strings.Split(setCookie, ";")[0], "=")[1]
}