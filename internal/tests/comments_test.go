package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

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
