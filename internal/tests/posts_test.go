package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIPostsList(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=10", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	// data is a list of posts
	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("unmarshal posts: %v", err)
	}
	if len(posts) == 0 {
		t.Errorf("expected at least 1 post from seed data, got 0")
	}

	// meta.total should be >= len(posts)
	var meta struct {
		Page    int `json:"page"`
		PerPage int `json:"per_page"`
		Total   int `json:"total"`
	}
	if len(env.Meta) > 0 {
		if err := json.Unmarshal(env.Meta, &meta); err != nil {
			t.Fatalf("unmarshal meta: %v", err)
		}
		if meta.Total < len(posts) {
			t.Errorf("expected meta.total >= len(posts), got total=%d, len=%d", meta.Total, len(posts))
		}
	}
}

func TestAPIPostsCreate(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register and login a user
	regBody := `{"username":"testuser2","email":"test2@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}
	loginBody := `{"username":"testuser2","password":"password123"}`
	req = httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}
	// Extract token from cookie
	setCookie := rec.Header().Get("Set-Cookie")
	parts := strings.Split(setCookie, ";")
	token := strings.TrimPrefix(strings.TrimSpace(parts[0]), "session_token=")

	payload := map[string]any{
		"title": "API Test Post",
		"body":  "Body from API test",
		// category_id optional; schema allows NULL
	}
	bodyBytes, _ := json.Marshal(payload)

	// POST with cookie
	req = httptest.NewRequest("POST", "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	body := w.Body.Bytes()
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var post struct {
		ID        int64  `json:"id"`
		Title     string `json:"title"`
		Body      string `json:"body"`
		AuthorID  int64  `json:"author_id"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}

	if post.ID == 0 {
		t.Errorf("expected non-zero ID")
	}
	if post.Title != payload["title"] {
		t.Errorf("expected title %q, got %q", payload["title"], post.Title)
	}
	if post.Body != payload["body"] {
		t.Errorf("expected body %q, got %q", payload["body"], post.Body)
	}
	if post.AuthorID != 1 {
		t.Errorf("expected AuthorID=1 (fake author), got %d", post.AuthorID)
	}
	if post.CreatedAt == "" {
		t.Errorf("expected CreatedAt to be set")
	}
}

func TestAPIPostsCreateWithCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO categories (id, name, slug, created_at)
		VALUES 
		(2, 'Cat2', 'cat2', strftime('%Y-%m-%dT%H:%M:%SZ','now')),
		(3, 'Cat3', 'cat3', strftime('%Y-%m-%dT%H:%M:%SZ','now'));
	`)
	if err != nil {
		t.Fatalf("failed to seed categories: %v", err)
	}

	loginBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie header on login")
	}
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// ---- CREATE POST with categories ----

	payload := map[string]any{
		"title":        "Post with categories",
		"body":         "Body here",
		"category_ids": []int64{1, 2, 3},
	}
	bodyBytes, _ := json.Marshal(payload)

	req = httptest.NewRequest("POST", "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	body := rec.Body.Bytes()

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	var post map[string]any
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}

	cids := post["category_ids"].([]interface{})
	if len(cids) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cids))
	}
}

func TestAPIPostGetReturnsCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Seed categories + post_categories
	_, err := db.Exec(`
		INSERT INTO categories (id, name, slug, created_at)
		VALUES 
		(2, 'Extra', 'extra', strftime('%Y-%m-%dT%H:%M:%SZ','now'));

		INSERT INTO post_categories (post_id, category_id) VALUES (1, 1);
		INSERT INTO post_categories (post_id, category_id) VALUES (1,2);
	`)
	if err != nil {
		t.Fatalf("seed error: %v", err)
	}

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts/1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	json.Unmarshal(body, &env)

	var post map[string]any
	json.Unmarshal(env.Data, &post)

	cats := post["category_ids"].([]interface{})
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
}
