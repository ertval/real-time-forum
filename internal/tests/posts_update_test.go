package tests

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAPIPostUpdate_OnlyAuthorCanUpdate(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// User A (seed user) creates the post
	tokenA := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, tokenA, map[string]any{
		"title":        "Author Title",
		"body":         "Author Body",
		"category_ids": []int64{1},
	})

	// Register + login User B
	registerUser(t, h, "otheruser", "other@example.com", "password123")
	tokenB := loginAndGetToken(t, h, "otheruser", "password123")

	// User B attempts to update User A's post
	rec := patchPost(t, h, tokenB, postID, map[string]any{
		"title": "Hacked Title",
	})

	if rec.Code != http.StatusForbidden && rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 or 403, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Verify post did not change
	post := getPost(t, h, postID)
	if post["title"] != "Author Title" {
		t.Fatalf("expected title unchanged, got %v", post["title"])
	}
	if post["body"] != "Author Body" {
		t.Fatalf("expected body unchanged, got %v", post["body"])
	}
}

func TestAPIPostUpdate_Content(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	cases := []struct {
		name         string
		patch        map[string]any
		wantTitle    string
		wantBody     string
		initialTitle string
		initialBody  string
	}{
		{
			name:         "title only",
			patch:        map[string]any{"title": "Updated Title"},
			initialTitle: "Original Title",
			initialBody:  "Original Body",
			wantTitle:    "Updated Title",
			wantBody:     "Original Body",
		},
		{
			name:         "body only",
			patch:        map[string]any{"body": "Updated Body"},
			initialTitle: "Original Title",
			initialBody:  "Original Body",
			wantTitle:    "Original Title",
			wantBody:     "Updated Body",
		},
		{
			name:         "title and body",
			patch:        map[string]any{"title": "Updated Title", "body": "Updated Body"},
			initialTitle: "Original Title",
			initialBody:  "Original Body",
			wantTitle:    "Updated Title",
			wantBody:     "Updated Body",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			postID := createPostAndGetID(t, h, token, map[string]any{
				"title":        tc.initialTitle,
				"body":         tc.initialBody,
				"category_ids": []int64{1},
			})

			rec := patchPost(t, h, token, postID, tc.patch)
			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
			}

			post := getPost(t, h, postID)
			if post["title"] != tc.wantTitle {
				t.Fatalf("expected title %q, got %v", tc.wantTitle, post["title"])
			}
			if post["body"] != tc.wantBody {
				t.Fatalf("expected body %q, got %v", tc.wantBody, post["body"])
			}
		})
	}
}

func TestAPIPostUpdate_Validation(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	cases := []struct {
		name  string
		patch map[string]any
	}{
		{name: "empty title", patch: map[string]any{"title": ""}},
		{name: "empty body", patch: map[string]any{"body": ""}},
		{name: "no fields", patch: map[string]any{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			postID := createPostAndGetID(t, h, token, map[string]any{
				"title":        "Valid Title",
				"body":         "Valid Body",
				"category_ids": []int64{1},
			})
			rec := patchPost(t, h, token, postID, tc.patch)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
			}

			// Ensure unchanged for all invalid update attempts.
			post := getPost(t, h, postID)
			if post["title"] != "Valid Title" || post["body"] != "Valid Body" {
				t.Fatalf("post should be unchanged, got %v", post)
			}
		})
	}
}

func TestAPIPostUpdate_NotFound(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	const missingID int64 = 999999

	rec := patchPost(t, h, token, missingID, map[string]any{
		"title": "Does not matter",
	})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Error == nil {
		t.Fatalf("expected error envelope, got none: %s", rec.Body.String())
	}

	// align with your app's error conventions
	if env.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected NOT_FOUND, got %q", env.Error.Code)
	}
	if env.Error.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestAPIPostUpdate_UpdatedAtIsSet(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	// Create post
	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Original Title",
		"body":         "Original Body",
		"category_ids": []int64{1},
	})

	// Fetch post before update
	postBefore := getPost(t, h, postID)

	updatedAtBeforeAny, ok := postBefore["updated_at"]
	if !ok {
		t.Fatalf("expected updated_at in post before update, got %v", postBefore)
	}
	updatedAtBefore, ok := updatedAtBeforeAny.(string)
	if !ok || updatedAtBefore == "" {
		t.Fatalf("expected updated_at to be non-empty string, got %T (%v)", updatedAtBeforeAny, updatedAtBeforeAny)
	}

	// Ensure timestamp changes (SQLite uses second precision)
	time.Sleep(1100 * time.Millisecond)

	// Update title
	rec := patchPost(t, h, token, postID, map[string]any{
		"title": "Updated Title",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	// Fetch post after update
	postAfter := getPost(t, h, postID)

	updatedAtAfterAny, ok := postAfter["updated_at"]
	if !ok {
		t.Fatalf("expected updated_at in post after update, got %v", postAfter)
	}
	updatedAtAfter, ok := updatedAtAfterAny.(string)
	if !ok || updatedAtAfter == "" {
		t.Fatalf("expected updated_at to be non-empty string, got %T (%v)", updatedAtAfterAny, updatedAtAfterAny)
	}

	if updatedAtAfter == updatedAtBefore {
		t.Fatalf("expected updated_at to change after update, before=%q after=%q",
			updatedAtBefore, updatedAtAfter)
	}
}
