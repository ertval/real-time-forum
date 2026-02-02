package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPIPostsCreate(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	registerUser(t, h, "testuser2", "test2@example.com", "password123")
	token := loginAndGetToken(t, h, "testuser2", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "API Test Post",
		"body":         "Body from API test",
		"category_ids": []int64{1},
	})

	post := getPost(t, h, postID)

	if post["title"] != "API Test Post" {
		t.Fatalf("expected title %q, got %v", "API Test Post", post["title"])
	}
	if post["body"] != "Body from API test" {
		t.Fatalf("expected body %q, got %v", "Body from API test", post["body"])
	}
	if post["author"] != "testuser2" {
		t.Fatalf("expected author %q, got %v", "testuser2", post["author"])
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

	rec := createPostRequestWithoutCookie(t, h, nil, map[string]any{
		"title":        "Unauthed post",
		"body":         "Body",
		"category_ids": []int64{1},
	})

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostsCreate_InvalidSessionToken(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	bad := "definitely-not-valid"
	rec := createPostRequestWithoutCookie(t, h, &bad, map[string]any{
		"title":        "Bad token post",
		"body":         "Body",
		"category_ids": []int64{1},
	})

	if rec.Code != http.StatusUnauthorized && rec.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostsCreate_CategoryIDsRequired(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	cases := []struct {
		name    string
		title   string
		payload map[string]any
	}{
		{
			name:  "missing category_ids",
			title: "Post missing category_ids",
			payload: map[string]any{
				"title": "Post missing category_ids",
				"body":  "Body here",
			},
		},
		{
			name:  "empty category_ids",
			title: "Post empty category_ids",
			payload: map[string]any{
				"title":        "Post empty category_ids",
				"body":         "Body here",
				"category_ids": []int64{},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rec := createPostRequest(t, h, token, tc.payload)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
			}

			if got := countPostsByTitle(t, db, tc.title); got != 0 {
				t.Fatalf("expected no post created, found %d", got)
			}
		})
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

	token := loginAndGetToken(t, h, "testuser", "password123")

	_ = createPostAndGetID(t, h, token, map[string]any{
		"title":        "Post with categories",
		"body":         "Body here",
		"category_ids": []int64{1, 2, 3},
	})
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
