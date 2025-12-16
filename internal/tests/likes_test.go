package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIPostsLikeToggle(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// ------------------------------------------------------------
	// Register user
	// ------------------------------------------------------------
	regBody := `{"username":"liker","email":"liker@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}

	// ------------------------------------------------------------
	// Login
	// ------------------------------------------------------------
	loginBody := `{"username":"liker","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie")
	}
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// ------------------------------------------------------------
	// First toggle → LIKE
	// ------------------------------------------------------------
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env1 apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env1)

	var resp1 struct {
		PostID int64 `json:"post_id"`
		Liked  bool  `json:"liked"`
	}
	json.Unmarshal(env1.Data, &resp1)

	if !resp1.Liked {
		t.Fatalf("expected liked=true after first toggle")
	}

	// ------------------------------------------------------------
	// Second toggle → UNLIKE
	// ------------------------------------------------------------
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env2 apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env2)

	var resp2 struct {
		PostID int64 `json:"post_id"`
		Liked  bool  `json:"liked"`
	}
	json.Unmarshal(env2.Data, &resp2)

	if resp2.Liked {
		t.Fatalf("expected liked=false after second toggle")
	}
}
