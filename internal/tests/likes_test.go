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
	// Login user
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
		t.Fatalf("expected Set-Cookie header on login")
	}
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// ------------------------------------------------------------
	// First call: like
	// ------------------------------------------------------------
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on first like, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var env1 apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env1); err != nil {
		t.Fatalf("unmarshal envelope 1: %v", err)
	}
	if env1.Error != nil {
		t.Fatalf("unexpected error on first like: %+v", env1.Error)
	}

	var resp1 struct {
		PostID int64 `json:"post_id"`
		Liked  bool  `json:"liked"`
		Likes  int   `json:"likes"`
	}
	if err := json.Unmarshal(env1.Data, &resp1); err != nil {
		t.Fatalf("unmarshal data 1: %v", err)
	}

	if resp1.PostID != 1 {
		t.Errorf("expected post_id=1, got %d", resp1.PostID)
	}
	if !resp1.Liked {
		t.Errorf("expected liked=true after first toggle")
	}
	if resp1.Likes < 1 {
		t.Errorf("expected likes >= 1 after first toggle, got %d", resp1.Likes)
	}

	// ------------------------------------------------------------
	// Second call: unlike
	// ------------------------------------------------------------
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 on second like (unlike), got %d, body=%s", rec.Code, rec.Body.String())
	}

	var env2 apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env2); err != nil {
		t.Fatalf("unmarshal envelope 2: %v", err)
	}
	if env2.Error != nil {
		t.Fatalf("unexpected error on second like: %+v", env2.Error)
	}

	var resp2 struct {
		PostID int64 `json:"post_id"`
		Liked  bool  `json:"liked"`
		Likes  int   `json:"likes"`
	}
	if err := json.Unmarshal(env2.Data, &resp2); err != nil {
		t.Fatalf("unmarshal data 2: %v", err)
	}

	if resp2.PostID != 1 {
		t.Errorf("expected post_id=1, got %d", resp2.PostID)
	}
	if resp2.Liked {
		t.Errorf("expected liked=false after second toggle")
	}
	if resp2.Likes > resp1.Likes {
		t.Errorf(
			"expected likes to stay same or decrease, got before=%d after=%d",
			resp1.Likes,
			resp2.Likes,
		)
	}
}
