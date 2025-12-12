package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPICommentsCreateAndList(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// --------------------------------------------------
	// Register + Login user
	// --------------------------------------------------
	regBody := `{"username":"commenter","email":"c@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"commenter","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie on login")
	}

	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// --------------------------------------------------
	// Create comment (AUTHENTICATED)
	// --------------------------------------------------
	createPayload := map[string]any{
		"body": "Comment from API test",
	}
	bodyBytes, _ := json.Marshal(createPayload)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/posts/1/comments",
		bytes.NewReader(bodyBytes),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	body := w.Body.Bytes()

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201 for comment create, got %d, body=%s", w.Code, string(body))
	}

	// --------------------------------------------------
	// Validate create response
	// --------------------------------------------------
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error on create: %+v", env.Error)
	}

	var comment struct {
		ID     int64  `json:"id"`
		PostID int64  `json:"post_id"`
		UserID int64  `json:"user_id"`
		Body   string `json:"body"`
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

	// --------------------------------------------------
	// List comments (public GET)
	// --------------------------------------------------
	w2, body2 := doRequest(
		t,
		h,
		http.MethodGet,
		"/api/v1/posts/1/comments?page=1&per_page=10",
		nil,
	)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200 for list, got %d, body=%s", w2.Code, string(body2))
	}

	var env2 apiEnvelope
	if err := json.Unmarshal(body2, &env2); err != nil {
		t.Fatalf("unmarshal envelope list: %v", err)
	}

	var comments []map[string]any
	if err := json.Unmarshal(env2.Data, &comments); err != nil {
		t.Fatalf("unmarshal comments: %v", err)
	}

	if len(comments) == 0 {
		t.Fatalf("expected at least 1 comment in list")
	}

	found := false
	for _, c := range comments {
		if c["body"] == createPayload["body"] {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected to find created comment body in list")
	}
}
