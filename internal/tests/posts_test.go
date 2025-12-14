package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ------------------------------------------------------------
// LIST POSTS (PUBLIC)
// ------------------------------------------------------------

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

	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("unmarshal posts: %v", err)
	}
	if len(posts) == 0 {
		t.Errorf("expected at least 1 post from seed data, got 0")
	}
}

// ------------------------------------------------------------
// CREATE POST (AUTH REQUIRED)
// ------------------------------------------------------------

func TestAPIPostsCreate(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register
	regBody := `{"username":"testuser2","email":"test2@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", rec.Code)
	}

	// Login
	loginBody := `{"username":"testuser2","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie header")
	}
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	payload := map[string]any{
		"title": "API Test Post",
		"body":  "Body from API test",
	}
	bodyBytes, _ := json.Marshal(payload)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	var post struct {
		ID       int64  `json:"id"`
		Title    string `json:"title"`
		Body     string `json:"body"`
		AuthorID int64  `json:"author_id"`
	}
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}

	if post.AuthorID != 2 {
		t.Errorf("expected AuthorID=2, got %d", post.AuthorID)
	}
}

// ------------------------------------------------------------
// CREATE POST WITH CATEGORIES (AUTH REQUIRED)
// ------------------------------------------------------------

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
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	token := strings.Split(strings.Split(rec.Header().Get("Set-Cookie"), ";")[0], "=")[1]

	payload := map[string]any{
		"title":        "Post with categories",
		"body":         "Body here",
		"category_ids": []int64{1, 2, 3},
	}
	bodyBytes, _ := json.Marshal(payload)

	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

// ------------------------------------------------------------
// GET POST RETURNS CATEGORIES (PUBLIC)
// ------------------------------------------------------------

func TestAPIPostGetReturnsCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO categories (id, name, slug, created_at)
		VALUES (2, 'Extra', 'extra', strftime('%Y-%m-%dT%H:%M:%SZ','now'));

		INSERT INTO post_categories (post_id, category_id) VALUES (1, 1);
		INSERT INTO post_categories (post_id, category_id) VALUES (1, 2);
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

	categoryIDs := post["category_ids"].([]any)
	if len(categoryIDs) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categoryIDs))
	}
}
