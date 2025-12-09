package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

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
