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
		ID       int64  `json:"id"`
		PostID   int64  `json:"post_id"`
		UserID   int64  `json:"user_id"`
		Body     string `json:"body"`
		Likes    int    `json:"likes"`
		Dislikes int    `json:"dislikes"`
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
	if comment.Likes != 0 || comment.Dislikes != 0 {
		t.Errorf("expected zero reactions on new comment")
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

			if _, ok := c["likes"]; !ok {
				t.Fatalf("expected likes field in comment")
			}
			if _, ok := c["dislikes"]; !ok {
				t.Fatalf("expected dislikes field in comment")
			}
			break
		}
	}
	if !found {
		t.Errorf("expected to find created comment body in list")
	}
}

func TestGuestCannotCreateComment(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	payload := []byte(`{"body":"hello"}`)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/posts/1/comments",
		bytes.NewReader(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGuestCanListComments(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, _ := doRequest(
		t,
		h,
		http.MethodGet,
		"/api/v1/posts/1/comments",
		nil,
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCommentGet(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Assume no comment exists
	w, body := doRequest(t, h, http.MethodGet, "/api/v1/comments/1", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d, body=%s", w.Code, string(body))
	}
}
	var createdComment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envCreate.Data, &createdComment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	commentID := createdComment.ID

	// Now get the comment
	w, body := doRequest(t, h, http.MethodGet, fmt.Sprintf("/api/v1/comments/%d", commentID), nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var comment map[string]any
	if err := json.Unmarshal(env.Data, &comment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	if comment["body"] != createPayload["body"] {
		t.Errorf("expected body %q, got %q", createPayload["body"], comment["body"])
	}
	if _, ok := comment["likes"]; !ok {
		t.Errorf("expected likes field")
	}
	if _, ok := comment["dislikes"]; !ok {
		t.Errorf("expected dislikes field")
	}
}

func TestCommentUpdate(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login as commenter
	regBody := `{"username":"updater","email":"u@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"updater","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// Create comment
	createPayload := map[string]any{"body": "Original comment"}
	bodyBytes, _ := json.Marshal(createPayload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", "session_token="+token)

	wCreate := httptest.NewRecorder()
	h.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("create comment failed: %d %s", wCreate.Code, wCreate.Body.String())
	}

	var envCreate apiEnvelope
	if err := json.Unmarshal(wCreate.Body.Bytes(), &envCreate); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	var createdComment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envCreate.Data, &createdComment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	commentID := createdComment.ID

	// Update comment
	payload := `{"body":"Updated comment"}`
	req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/comments/%d", commentID), bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var env apiEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}
	// Verify updated body
	var comment map[string]any
	if err := json.Unmarshal(env.Data, &comment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	if comment["body"] != "Updated comment" {
		t.Errorf("expected updated body")
	}
}

func TestCommentDelete(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login as deleter
	regBody := `{"username":"deleter","email":"d@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"deleter","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// Create comment
	createPayload := map[string]any{"body": "Comment to delete"}
	bodyBytes, _ := json.Marshal(createPayload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", "session_token="+token)

	wCreate := httptest.NewRecorder()
	h.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("create comment failed: %d %s", wCreate.Code, wCreate.Body.String())
	}

	var envCreate apiEnvelope
	if err := json.Unmarshal(wCreate.Body.Bytes(), &envCreate); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	var createdComment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envCreate.Data, &createdComment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	commentID := createdComment.ID

	// Delete comment
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/comments/%d", commentID), nil)
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCommentLike(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login
	regBody := `{"username":"liker","email":"l@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"liker","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// Create comment
	createPayload := map[string]any{"body": "Comment to like"}
	bodyBytes, _ := json.Marshal(createPayload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", "session_token="+token)

	wCreate := httptest.NewRecorder()
	h.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("create comment failed: %d %s", wCreate.Code, wCreate.Body.String())
	}

	var envCreate apiEnvelope
	if err := json.Unmarshal(wCreate.Body.Bytes(), &envCreate); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	var createdComment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envCreate.Data, &createdComment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	commentID := createdComment.ID

	// Like comment
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/comments/%d/like", commentID), nil)
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var env apiEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}
	// Verify reaction and counts
	var result map[string]any
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result["comment_id"] != float64(commentID) {
		t.Errorf("expected comment_id %d", commentID)
	}
	if result["reaction"] != float64(1) {
		t.Errorf("expected reaction 1")
	}
	if result["likes_count"] != float64(1) {
		t.Errorf("expected likes_count 1")
	}
}

func TestCommentDislike(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login
	regBody := `{"username":"disliker","email":"dl@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"disliker","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	// Create comment
	createPayload := map[string]any{"body": "Comment to dislike"}
	bodyBytes, _ := json.Marshal(createPayload)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/comments", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Cookie", "session_token="+token)

	wCreate := httptest.NewRecorder()
	h.ServeHTTP(wCreate, reqCreate)

	if wCreate.Code != http.StatusCreated {
		t.Fatalf("create comment failed: %d %s", wCreate.Code, wCreate.Body.String())
	}

	var envCreate apiEnvelope
	if err := json.Unmarshal(wCreate.Body.Bytes(), &envCreate); err != nil {
		t.Fatalf("unmarshal create: %v", err)
	}
	var createdComment struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(envCreate.Data, &createdComment); err != nil {
		t.Fatalf("unmarshal comment: %v", err)
	}
	commentID := createdComment.ID

	// Dislike comment
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/comments/%d/dislike", commentID), nil)
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var env apiEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}
	// Verify reaction and counts
	var result map[string]any
	if err := json.Unmarshal(env.Data, &result); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if result["comment_id"] != float64(commentID) {
		t.Errorf("expected comment_id %d", commentID)
	}
	if result["reaction"] != float64(-1) {
		t.Errorf("expected reaction -1")
	}
	if result["dislikes_count"] != float64(1) {
		t.Errorf("expected dislikes_count 1")
	}
}

func TestGuestCannotUpdateComment(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	payload := []byte(`{"body":"update"}`)
	w, _ := doRequest(t, h, http.MethodPatch, "/api/v1/comments/1", payload)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestGuestCannotDeleteComment(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, _ := doRequest(t, h, http.MethodDelete, "/api/v1/comments/1", nil)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestUserCannotUpdateOthersComment(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login as non-owner
	regBody := `{"username":"other","email":"o@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"other","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	payload := []byte(`{"body":"update"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/comments/1", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestUserCannotDeleteOthersComment(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Register/login as non-owner
	regBody := `{"username":"other2","email":"o2@example.com","password":"password123"}`
	_, _ = doRequest(t, h, http.MethodPost, "/api/v1/users/register", []byte(regBody))

	loginBody := `{"username":"other2","password":"password123"}`
	rec, _ := doRequest(t, h, http.MethodPost, "/api/v1/users/login", []byte(loginBody))
	setCookie := rec.Header().Get("Set-Cookie")
	token := strings.Split(strings.Split(setCookie, ";")[0], "=")[1]

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/comments/1", nil)
	req.Header.Set("Cookie", "session_token="+token)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
