package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	regBody := `{"username":"otheruser","email":"other@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register other user failed: %d body=%s", rec.Code, rec.Body.String())
	}

	tokenB := loginAndGetToken(t, h, "otheruser", "password123")

	// User B attempts to update User A's post
	rec = patchPost(t, h, tokenB, postID, map[string]any{
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

func TestAPIPostUpdate_TitleOnly(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Original Title",
		"body":         "Original Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{
		"title": "Updated Title",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	post := getPost(t, h, postID)

	if post["title"] != "Updated Title" {
		t.Fatalf("expected updated title, got %v", post["title"])
	}
	if post["body"] != "Original Body" {
		t.Fatalf("expected body unchanged, got %v", post["body"])
	}
}

func TestAPIPostUpdate_BodyOnly(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Original Title",
		"body":         "Original Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{
		"body": "Updated Body",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	post := getPost(t, h, postID)

	if post["title"] != "Original Title" {
		t.Fatalf("expected title unchanged, got %v", post["title"])
	}
	if post["body"] != "Updated Body" {
		t.Fatalf("expected updated body, got %v", post["body"])
	}
}

func TestAPIPostUpdate_TitleAndBody(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Original Title",
		"body":         "Original Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{
		"title": "Updated Title",
		"body":  "Updated Body",
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	post := getPost(t, h, postID)

	if post["title"] != "Updated Title" {
		t.Fatalf("expected updated title, got %v", post["title"])
	}
	if post["body"] != "Updated Body" {
		t.Fatalf("expected updated body, got %v", post["body"])
	}
}

func TestAPIPostUpdate_EmptyTitle(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Valid Title",
		"body":         "Valid Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{
		"title": "",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostUpdate_EmptyBody(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Valid Title",
		"body":         "Valid Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{
		"body": "",
	})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAPIPostUpdate_NoFields(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	postID := createPostAndGetID(t, h, token, map[string]any{
		"title":        "Original Title",
		"body":         "Original Body",
		"category_ids": []int64{1},
	})

	rec := patchPost(t, h, token, postID, map[string]any{})

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	post := getPost(t, h, postID)
	if post["title"] != "Original Title" || post["body"] != "Original Body" {
		t.Fatalf("post should be unchanged, got %v", post)
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

/*---------
  HELPERS
---------*/

func createPostAndGetID(t *testing.T, h http.Handler, token string, payload map[string]any) int64 {
	t.Helper()

	rec := createPostRequest(t, h, token, payload)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var post map[string]any
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}

	idAny, ok := post["id"]
	if !ok {
		t.Fatalf("expected id in create response, post=%v", post)
	}
	return int64(idAny.(float64))
}

func patchPost(t *testing.T, h http.Handler, token string, postID int64, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/posts/%d", postID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func getPost(t *testing.T, h http.Handler, postID int64) map[string]any {
	t.Helper()

	w, body := doRequest(t, h, http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", postID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var post map[string]any
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}
	return post
}
