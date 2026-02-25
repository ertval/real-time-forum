// internal/tests/notifications_api_test.go
package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	repository "forum/internal/db"
)

type notificationsResponse struct {
	Data []any `json:"data"`
	Meta struct {
		UnreadCount int `json:"unread_count"`
	} `json:"meta"`
}

func TestNotifications_AuthRequired(t *testing.T) {
	h, _ := newTestAPI(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestNotifications_GetAndUnreadCount(t *testing.T) {
	h, database := newTestAPI(t)
	ctx := context.Background()

	author := createTestUser(t, database, "author")
	user := createTestUser(t, database, "user")
	post := createTestPost(t, database, author)

	// user likes author's post
	_, err := repository.ToggleReaction(ctx, database, user, post, 1, "post")
	if err != nil {
		t.Fatalf("toggle reaction: %v", err)
	}

	// LOGIN AS AUTHOR (recipient!)
	token := loginTestUser(t, h, "author")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp notificationsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	if resp.Meta.UnreadCount != 1 {
		t.Fatalf("expected unread_count 1, got %d", resp.Meta.UnreadCount)
	}
}

func TestNotifications_ReadOne(t *testing.T) {
	h, database := newTestAPI(t)
	ctx := context.Background()

	author := createTestUser(t, database, "author")
	user := createTestUser(t, database, "user")
	post := createTestPost(t, database, author)

	_, _ = repository.ToggleReaction(ctx, database, user, post, 1, "post")

	token := loginTestUser(t, h, "author")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/1/read", nil)
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNotifications_ReadAll(t *testing.T) {
	h, database := newTestAPI(t)
	ctx := context.Background()

	author := createTestUser(t, database, "author")
	user := createTestUser(t, database, "user")
	post := createTestPost(t, database, author)

	_, _ = repository.ToggleReaction(ctx, database, user, post, 1, "post")

	token := loginTestUser(t, h, "author")

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/read-all", nil)
	req.Header.Set("Cookie", "session_token="+token)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
