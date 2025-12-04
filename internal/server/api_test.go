package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"forum/internal/router"
)

// ---------- Test helpers ----------

// setupTestDB creates an in-memory SQLite DB, loads forum_schema.sql, and seeds minimal data.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// forum_schema.sql is in internal/db relative to this package (internal/server)
	schemaPath := filepath.Join("..", "db", "forum_schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		db.Close()
		t.Fatalf("failed to read schema file %s: %v", schemaPath, err)
	}

	if _, err := db.Exec(string(schemaBytes)); err != nil {
		db.Close()
		t.Fatalf("failed to exec schema: %v", err)
	}

	// Seed one user, one category, one post
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at)
		VALUES (1, 'testuser', 'test@example.com', 'hash', 1, datetime('now'), datetime('now'));

		INSERT INTO categories (id, name, slug, created_at)
		VALUES (1, 'Test Category', 'test-category', datetime('now'));

		INSERT INTO posts (id, author_id, title, body, status, category_id, created_at, updated_at)
		VALUES (1, 1, 'Seed Post', 'Seed post body', 'published', 1, datetime('now'), datetime('now'));
	`)
	if err != nil {
		db.Close()
		t.Fatalf("failed to seed initial data: %v", err)
	}

	return db
}

// newTestAPI builds the HTTP handler (router + middleware) using an in-memory DB.
func newTestAPI(t *testing.T) (http.Handler, *sql.DB) {
	t.Helper()
	db := setupTestDB(t)
	h := router.NewRouter(db)
	return h, db
}

// small helper to perform a request and return recorder + body bytes.
func doRequest(t *testing.T, h http.Handler, method, path string, body []byte) (*httptest.ResponseRecorder, []byte) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req = req.WithContext(context.Background())
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	respBody := w.Body.Bytes()
	return w, respBody
}

// generic envelope used by API
type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Meta  json.RawMessage `json:"meta,omitempty"`
	Error *apiError       `json:"error,omitempty"`
}

// ---------- Tests ----------

func TestAPIHealth(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/health", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error in response: %+v", env.Error)
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(env.Data, &payload); err != nil {
		t.Fatalf("failed to unmarshal data: %v", err)
	}
	if payload.Status != "ok" {
		t.Errorf("expected status 'ok', got %q", payload.Status)
	}
}

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

	payload := map[string]any{
		"title": "API Test Post",
		"body":  "Body from API test",
		// category_id optional; schema allows NULL
	}
	bodyBytes, _ := json.Marshal(payload)

	w, body := doRequest(t, h, http.MethodPost, "/api/v1/posts", bodyBytes)
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

func TestAPICommentsCreateAndList(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// First create a comment on post 1
	createPayload := map[string]any{
		"body": "Comment from API test",
	}
	bodyBytes, _ := json.Marshal(createPayload)

	commentURL := "/api/v1/posts/1/comments"

	w, body := doRequest(t, h, http.MethodPost, commentURL, bodyBytes)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for comment create, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error on create: %+v", env.Error)
	}

	var comment struct {
		ID        int64  `json:"id"`
		PostID    int64  `json:"post_id"`
		UserID    int64  `json:"user_id"`
		Body      string `json:"body"`
		CreatedAt string `json:"created_at"`
	}
	if err := json.Unmarshal(env.Data, &comment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	if comment.ID == 0 {
		t.Errorf("expected non-zero comment ID")
	}
	if comment.PostID != 1 {
		t.Errorf("expected PostID=1, got %d", comment.PostID)
	}
	if comment.Body != createPayload["body"] {
		t.Errorf("expected body %q, got %q", createPayload["body"], comment.Body)
	}

	// Now list comments for post 1 and ensure our comment appears
	w2, body2 := doRequest(t, h, http.MethodGet, "/api/v1/posts/1/comments?page=1&per_page=10", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list, got %d, body=%s", w2.Code, string(body2))
	}

	var env2 apiEnvelope
	if err := json.Unmarshal(body2, &env2); err != nil {
		t.Fatalf("unmarshal envelope list: %v", err)
	}
	if env2.Error != nil {
		t.Fatalf("unexpected error on list: %+v", env2.Error)
	}

	var comments []map[string]any
	if err := json.Unmarshal(env2.Data, &comments); err != nil {
		t.Fatalf("unmarshal comments: %v", err)
	}
	if len(comments) == 0 {
		t.Fatalf("expected at least 1 comment in list, got 0")
	}

	// Optional: quick check one of them matches our created body
	found := false
	for _, c := range comments {
		if bodyVal, ok := c["body"].(string); ok && bodyVal == createPayload["body"] {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find created comment body %q in list", createPayload["body"])
	}
}

func TestAPIPostsLikeToggle(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// First call: should create a like
	w1, body1 := doRequest(t, h, http.MethodPost, "/api/v1/posts/1/like", nil)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on first like, got %d, body=%s", w1.Code, string(body1))
	}

	var env1 apiEnvelope
	if err := json.Unmarshal(body1, &env1); err != nil {
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

	// Second call: should unlike
	w2, body2 := doRequest(t, h, http.MethodPost, "/api/v1/posts/1/like", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on second like (unlike), got %d, body=%s", w2.Code, string(body2))
	}

	var env2 apiEnvelope
	if err := json.Unmarshal(body2, &env2); err != nil {
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
		t.Errorf("expected likes to stay same or decrease, got before=%d after=%d", resp1.Likes, resp2.Likes)
	}
}

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
	// Test user retrieval
	userID := int64(id.(float64))
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/%d", userID), nil)
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
	var loginResp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}
	userData, ok := loginResp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data in login response")
	}
	token, ok := userData["token"].(string) // Assuming login returns token
	if !ok {
		t.Fatalf("expected token in login response")
	}
	// Call /me with token
	req = httptest.NewRequest("GET", "/api/v1/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
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
