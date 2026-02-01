package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
		"title":        "API Test Post",
		"body":         "Body from API test",
		"category_ids": []int64{1},
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

func TestAPIPostsCreate_InvalidTitleOrBody(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	cases := []struct {
		name    string
		payload map[string]any
	}{
		{
			name: "missing title",
			payload: map[string]any{
				"body":         "Body",
				"category_ids": []int64{1},
			},
		},
		{
			name: "empty title",
			payload: map[string]any{
				"title":        "",
				"body":         "Body",
				"category_ids": []int64{1},
			},
		},
		{
			name: "missing body",
			payload: map[string]any{
				"title":        "Title",
				"category_ids": []int64{1},
			},
		},
		{
			name: "empty body",
			payload: map[string]any{
				"title":        "Title",
				"body":         "",
				"category_ids": []int64{1},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := createPostRequest(t, h, token, tc.payload)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAPIPostsCreate_RequiresAuth(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	payload := map[string]any{
		"title":        "Unauthed post",
		"body":         "Body",
		"category_ids": []int64{1},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// no Cookie header

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostsCreate_InvalidSessionToken(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	payload := map[string]any{
		"title":        "Bad token post",
		"body":         "Body",
		"category_ids": []int64{1},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token=definitely-not-valid")

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostsCreate_MissingCategoryIDs(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	title := "Post missing category_ids"
	rec := createPostRequest(t, h, token, map[string]any{
		"title": title,
		"body":  "Body here",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	if got := countPostsByTitle(t, db, title); got != 0 {
		t.Fatalf("expected no post created, found %d", got)
	}
}

func TestAPIPostsCreate_EmptyCategoryIDs(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	title := "Post empty category_ids"
	rec := createPostRequest(t, h, token, map[string]any{
		"title":        title,
		"body":         "Body here",
		"category_ids": []int64{},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	if got := countPostsByTitle(t, db, title); got != 0 {
		t.Fatalf("expected no post created, found %d", got)
	}
}

func TestAPIPostsCreate_DuplicateCategoryIDs(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// ensure category 2 exists (1 likely exists in seed)
	_, err := db.Exec(`
		INSERT INTO categories (id, name, created_at)
		VALUES (2, 'Cat2', strftime('%Y-%m-%dT%H:%M:%SZ','now'))
	`)
	if err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	token := loginAndGetToken(t, h, "testuser", "password123")

	title := "Post with dup cats"
	rec := createPostRequest(t, h, token, map[string]any{
		"title":        title,
		"body":         "Body here",
		"category_ids": []int64{1, 1, 2, 2, 2},
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Expect only 2 join rows (category 1 and 2 once each)
	if got := countPostCategoriesForTitle(t, db, title); got != 2 {
		t.Fatalf("expected 2 unique post_categories rows, got %d", got)
	}
}

func TestAPIPostsCreate_InvalidCategoryIDs(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	title := "Post invalid cats only"
	rec := createPostRequest(t, h, token, map[string]any{
		"title":        title,
		"body":         "Body here",
		"category_ids": []int64{0, -1, -50},
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	if got := countPostsByTitle(t, db, title); got != 0 {
		t.Fatalf("expected no post created, found %d", got)
	}
}

func TestAPIPostsCreateWithCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO categories (id, name, created_at)
		VALUES 
		(2, 'Cat2', strftime('%Y-%m-%dT%H:%M:%SZ','now')),
		(3, 'Cat3', strftime('%Y-%m-%dT%H:%M:%SZ','now'));
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

func TestAPIPostsCreate_InvalidJSON(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	// deliberately broken JSON
	badJSON := `{"title": "X", "body": "Y", "category_ids": [1]`

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewBufferString(badJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Error == nil {
		t.Fatalf("expected error envelope, got none: %s", rec.Body.String())
	}
	if env.Error.Code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST, got %q", env.Error.Code)
	}
	if env.Error.Message != "invalid json" {
		t.Fatalf("expected message 'invalid json', got %q", env.Error.Message)
	}
}

/*---------
  HELPERS
----------*/

func countPostsByTitle(t *testing.T, db *sql.DB, title string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM posts WHERE title = ?`, title).Scan(&n); err != nil {
		t.Fatalf("count posts by title: %v", err)
	}
	return n
}

func countPostCategoriesForTitle(t *testing.T, db *sql.DB, title string) int {
	t.Helper()

	var n int
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM post_categories pc
		JOIN posts p ON p.id = pc.post_id
		WHERE p.title = ?
	`, title).Scan(&n); err != nil {
		t.Fatalf("count post_categories: %v", err)
	}
	return n
}

func createPostRequest(t *testing.T, h http.Handler, token string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}
