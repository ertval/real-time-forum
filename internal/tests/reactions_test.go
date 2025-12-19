package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type reactionResp struct {
	PostID        int64 `json:"post_id,omitempty"`
	CommentID     int64 `json:"comment_id,omitempty"`
	Reaction      int   `json:"reaction"` // 1, -1, or 0
	LikesCount    int   `json:"likes_count"`
	DislikesCount int   `json:"dislikes_count"`
}

func registerAndLogin(t *testing.T, h http.Handler) string {
	t.Helper()

	// Register
	regBody := `{"username":"reactor","email":"reactor@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("register failed: %d body=%s", rec.Code, rec.Body.String())
	}

	// Login
	loginBody := `{"username":"reactor","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d body=%s", rec.Code, rec.Body.String())
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie")
	}
	return strings.Split(strings.Split(setCookie, ";")[0], "=")[1]
}

func doPostReaction(t *testing.T, h http.Handler, token, url string) reactionResp {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, url, nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d url=%s body=%s", rec.Code, url, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v body=%s", err, rec.Body.String())
	}

	var resp reactionResp
	if err := json.Unmarshal(env.Data, &resp); err != nil {
		t.Fatalf("unmarshal response: %v data=%s", err, string(env.Data))
	}

	return resp
}

func createComment(t *testing.T, h http.Handler, token string, postID int64, body string) int64 {
	t.Helper()

	reqBody := map[string]string{"body": body}
	jsonBody, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/posts/"+fmt.Sprintf("%d", postID)+"/comments", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create comment failed: %d body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	var comment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(env.Data, &comment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}

	return comment.ID
}

func TestAPIPostsReactionToggle_LikeThenUnlike(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := registerAndLogin(t, h)

	// 1) like
	r1 := doPostReaction(t, h, token, "/api/v1/posts/1/like")
	if r1.PostID != 1 {
		t.Fatalf("expected post_id=1 got %d", r1.PostID)
	}
	if r1.Reaction != 1 {
		t.Fatalf("expected reaction=1 after like, got %d", r1.Reaction)
	}
	if r1.LikesCount != 1 || r1.DislikesCount != 0 {
		t.Fatalf("expected likes=1 dislikes=0 got likes=%d dislikes=%d", r1.LikesCount, r1.DislikesCount)
	}

	// 2) unlike (toggle off)
	r2 := doPostReaction(t, h, token, "/api/v1/posts/1/like")
	if r2.Reaction != 0 {
		t.Fatalf("expected reaction=0 after unlike, got %d", r2.Reaction)
	}
	if r2.LikesCount != 0 || r2.DislikesCount != 0 {
		t.Fatalf("expected likes=0 dislikes=0 got likes=%d dislikes=%d", r2.LikesCount, r2.DislikesCount)
	}
}

func TestAPIPostsReactionToggle_DislikeThenUndislike(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := registerAndLogin(t, h)

	// 1) dislike
	r1 := doPostReaction(t, h, token, "/api/v1/posts/1/dislike")
	if r1.PostID != 1 {
		t.Fatalf("expected post_id=1 got %d", r1.PostID)
	}
	if r1.Reaction != -1 {
		t.Fatalf("expected reaction=-1 after dislike, got %d", r1.Reaction)
	}
	if r1.LikesCount != 0 || r1.DislikesCount != 1 {
		t.Fatalf("expected likes=0 dislikes=1 got likes=%d dislikes=%d", r1.LikesCount, r1.DislikesCount)
	}

	// 2) undislike (toggle off)
	r2 := doPostReaction(t, h, token, "/api/v1/posts/1/dislike")
	if r2.Reaction != 0 {
		t.Fatalf("expected reaction=0 after undislike, got %d", r2.Reaction)
	}
	if r2.LikesCount != 0 || r2.DislikesCount != 0 {
		t.Fatalf("expected likes=0 dislikes=0 got likes=%d dislikes=%d", r2.LikesCount, r2.DislikesCount)
	}
}

func TestAPIPostsReactionToggle_SwitchLikeToDislike(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := registerAndLogin(t, h)

	// like first
	r1 := doPostReaction(t, h, token, "/api/v1/posts/1/like")
	if r1.PostID != 1 {
		t.Fatalf("expected post_id=1 got %d", r1.PostID)
	}
	if r1.Reaction != 1 || r1.LikesCount != 1 || r1.DislikesCount != 0 {
		t.Fatalf("expected reaction=1 likes=1 dislikes=0 got reaction=%d likes=%d dislikes=%d",
			r1.Reaction, r1.LikesCount, r1.DislikesCount)
	}

	// then switch to dislike
	r2 := doPostReaction(t, h, token, "/api/v1/posts/1/dislike")
	if r2.Reaction != -1 {
		t.Fatalf("expected reaction=-1 after switching to dislike, got %d", r2.Reaction)
	}
	if r2.LikesCount != 0 || r2.DislikesCount != 1 {
		t.Fatalf("expected likes=0 dislikes=1 after switch got likes=%d dislikes=%d", r2.LikesCount, r2.DislikesCount)
	}

	// clicking dislike again should remove it
	r3 := doPostReaction(t, h, token, "/api/v1/posts/1/dislike")
	if r3.Reaction != 0 || r3.LikesCount != 0 || r3.DislikesCount != 0 {
		t.Fatalf("expected reaction=0 likes=0 dislikes=0 after toggle off got reaction=%d likes=%d dislikes=%d",
			r3.Reaction, r3.LikesCount, r3.DislikesCount)
	}
}
