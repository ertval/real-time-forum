package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

/*------------
   GET POST
-------------*/

func TestAPIPostGetReturnsCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	_, err := db.Exec(`
		INSERT INTO categories (id, name, created_at)
		VALUES (2, 'Extra', strftime('%Y-%m-%dT%H:%M:%SZ','now'));

		INSERT INTO post_categories (post_id, category_id) VALUES (1, 1);
		INSERT INTO post_categories (post_id, category_id) VALUES (1, 2);
	`)
	if err != nil {
		t.Fatalf("seed error: %v", err)
	}

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts/1", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	json.Unmarshal(body, &env)

	var post map[string]any
	json.Unmarshal(env.Data, &post)

	raw, ok := post["categories"]
	if !ok || raw == nil {
		t.Fatalf("expected category_ids in response, got: %v", post)
	}

	cats, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected categories to be array, got %T (%v)", raw, raw)
	}

	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
}

func TestAPIPostGet_NotFound(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// pick an id that won't exist (test DB is tiny)
	const missingID = 9999

	w, body := doRequest(t, h, http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", missingID), nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	if env.Error == nil {
		t.Fatalf("expected error envelope, got none: %s", string(body))
	}

	if env.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected error code NOT_FOUND, got %q", env.Error.Code)
	}

	if env.Error.Message == "" {
		t.Fatalf("expected non-empty error message")
	}
}

func TestAPIPostGet_ReactionCounts(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Create a known post
	res, err := db.Exec(`
		INSERT INTO posts (author_id, title, body, status, created_at, updated_at)
		VALUES (1, 'ReactCount', 'Body', 'published', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	postID, _ := res.LastInsertId()

	_, err = db.Exec(`
		INSERT INTO users (username, email, password_hash, is_active, created_at, updated_at)
		VALUES
			('react_get_u1', 'react_get_u1@example.com', 'x', 1, datetime('now'), datetime('now')),
			('react_get_u2', 'react_get_u2@example.com', 'x', 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}

	var u1, u2 int64
	if err := db.QueryRow(`SELECT id FROM users WHERE username='react_get_u1'`).Scan(&u1); err != nil {
		t.Fatalf("get u1: %v", err)
	}
	if err := db.QueryRow(`SELECT id FROM users WHERE username='react_get_u2'`).Scan(&u2); err != nil {
		t.Fatalf("get u2: %v", err)
	}

	_, err = db.Exec(`INSERT INTO reactions (user_id, post_id, value, created_at) VALUES (?, ?, ?, datetime('now'))`, u1, postID, 1)
	if err != nil {
		t.Fatalf("seed like: %v", err)
	}
	_, err = db.Exec(`INSERT INTO reactions (user_id, post_id, value, created_at) VALUES (?, ?, ?, datetime('now'))`, u2, postID, -1)
	if err != nil {
		t.Fatalf("seed dislike: %v", err)
	}

	w, body := doRequest(t, h, http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", postID), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var post map[string]any
	if err := json.Unmarshal(env.Data, &post); err != nil {
		t.Fatalf("unmarshal post: %v", err)
	}

	likesAny, ok := post["likes"]
	if !ok {
		t.Fatalf("expected likes field, post=%v", post)
	}
	dislikesAny, ok := post["dislikes"]
	if !ok {
		t.Fatalf("expected dislikes field, post=%v", post)
	}

	likes := int(likesAny.(float64))
	dislikes := int(dislikesAny.(float64))

	if likes != 1 || dislikes != 1 {
		t.Fatalf("expected likes=1 dislikes=1, got likes=%d dislikes=%d post=%v", likes, dislikes, post)
	}
}

func TestAPIPostsFilterByCategory(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, body := doRequest(
		t,
		h,
		http.MethodGet,
		"/api/v1/posts?category_id=1",
		nil,
	)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", w.Code, string(body))
	}
}

// ------------------------------------------------------------
// LIST LIKED POSTS (AUTH REQUIRED)
// ------------------------------------------------------------

func TestAPIPostsLiked_RequiresAuth(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, _ := doRequest(t, h, http.MethodGet, "/api/v1/posts/liked", nil)

	if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
		t.Fatalf("expected 401 or 403, got %d", w.Code)
	}
}

func TestAPIPostsLiked_ReturnsOnlyLikedPosts(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// login as seed user
	loginBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d", rec.Code)
	}

	token := strings.Split(strings.Split(rec.Header().Get("Set-Cookie"), ";")[0], "=")[1]

	// like post id=1
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("like failed: %d", rec.Code)
	}

	// list liked posts
	req = httptest.NewRequest(http.MethodGet, "/api/v1/posts/liked", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}

	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("unmarshal posts: %v", err)
	}

	if len(posts) != 1 {
		t.Fatalf("expected 1 liked post, got %d", len(posts))
	}

	if int64(posts[0]["id"].(float64)) != 1 {
		t.Fatalf("expected post id=1")
	}
}

func TestAPIPostsLiked_UnlikeRemovesPost(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	loginBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	token := strings.Split(strings.Split(rec.Header().Get("Set-Cookie"), ";")[0], "=")[1]

	// like
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// unlike
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/like", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// list liked posts
	req = httptest.NewRequest(http.MethodGet, "/api/v1/posts/liked", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var env apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env)

	var posts []any
	json.Unmarshal(env.Data, &posts)

	if len(posts) != 0 {
		t.Fatalf("expected 0 liked posts, got %d", len(posts))
	}
}

func TestAPIPostsLiked_DislikeDoesNotCount(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	loginBody := `{"username":"testuser","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	token := strings.Split(strings.Split(rec.Header().Get("Set-Cookie"), ";")[0], "=")[1]

	// dislike post
	req = httptest.NewRequest(http.MethodPost, "/api/v1/posts/1/dislike", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	// list liked posts
	req = httptest.NewRequest(http.MethodGet, "/api/v1/posts/liked", nil)
	req.Header.Set("Cookie", "session_token="+token)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var env apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env)

	var posts []any
	json.Unmarshal(env.Data, &posts)

	if len(posts) != 0 {
		t.Fatalf("disliked post must not appear in liked posts")
	}
}

/*----------
 LIST POSTS
------------*/

// LIST POSTS (PUBLIC)
func TestAPIPostsList(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=10", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, string(body))
	}

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("unmarshal posts: %v", err)
	}
	if len(posts) == 0 {
		t.Errorf("expected at least 1 post from seed data, got 0")
	}
}

func TestAPIPostsList_Pagination(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	seedPosts(t, db, 1, 25) // enough to paginate

	// page=1 per_page=10 => 10 items
	w1, b1 := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=10", nil)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w1.Code, string(b1))
	}
	p1 := decodePostsList(t, b1)
	if len(p1) != 10 {
		t.Fatalf("expected 10 posts, got %d", len(p1))
	}

	// page=2 per_page=10 => 10 items, different ids from page 1
	w2, b2 := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=2&per_page=10", nil)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w2.Code, string(b2))
	}
	p2 := decodePostsList(t, b2)
	if len(p2) != 10 {
		t.Fatalf("expected 10 posts, got %d", len(p2))
	}

	ids1 := map[int64]struct{}{}
	for _, p := range p1 {
		id := int64(p["id"].(float64))
		ids1[id] = struct{}{}
	}
	for _, p := range p2 {
		id := int64(p["id"].(float64))
		if _, ok := ids1[id]; ok {
			t.Fatalf("expected page 2 not to overlap with page 1, but found id=%d", id)
		}
	}
}

func TestAPIPostsList_Defaulting(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	seedPosts(t, db, 1, 25)

	// page=0 should behave like page=1
	wBadPage, bBadPage := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=0&per_page=10", nil)
	if wBadPage.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", wBadPage.Code, string(bBadPage))
	}
	badPagePosts := decodePostsList(t, bBadPage)

	wPage1, bPage1 := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=10", nil)
	if wPage1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", wPage1.Code, string(bPage1))
	}
	page1Posts := decodePostsList(t, bPage1)

	if len(badPagePosts) != len(page1Posts) {
		t.Fatalf("expected same length for page=0 and page=1, got %d vs %d", len(badPagePosts), len(page1Posts))
	}
	// compare first id
	if int64(badPagePosts[0]["id"].(float64)) != int64(page1Posts[0]["id"].(float64)) {
		t.Fatalf("expected page=0 to default to page=1; first ids differ")
	}

	// per_page=0 should behave like per_page=20
	wBadPer, bBadPer := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=0", nil)
	if wBadPer.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", wBadPer.Code, string(bBadPer))
	}
	badPerPosts := decodePostsList(t, bBadPer)
	if len(badPerPosts) != 20 {
		t.Fatalf("expected per_page=0 to default to 20, got %d", len(badPerPosts))
	}
}

func TestAPIPostsList_Bounds(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	seedPosts(t, db, 1, 150) // ensure there are more than 100 posts total

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=1000", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}

	posts := decodePostsList(t, body)
	if len(posts) != 100 {
		t.Fatalf("expected per_page to clamp to 100, got %d", len(posts))
	}
}

func TestAPIPostsList_AttachesCategories(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Ensure categories exist
	_, err := db.Exec(`
		INSERT INTO categories (id, name, created_at)
		VALUES
			(101, 'Cat101', strftime('%Y-%m-%dT%H:%M:%SZ','now')),
			(102, 'Cat102', strftime('%Y-%m-%dT%H:%M:%SZ','now'))
	`)
	if err != nil {
		t.Fatalf("seed categories: %v", err)
	}

	// Create a known post
	res, err := db.Exec(`
		INSERT INTO posts (author_id, title, body, status, created_at, updated_at)
		VALUES (1, 'ListCats', 'Body', 'published', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	postID, _ := res.LastInsertId()

	// Attach categories
	_, err = db.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, 101)
	if err != nil {
		t.Fatalf("seed post_categories 101: %v", err)
	}
	_, err = db.Exec(`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`, postID, 102)
	if err != nil {
		t.Fatalf("seed post_categories 102: %v", err)
	}

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}
	posts := decodePostsList(t, body)

	// Find our post
	var got map[string]any
	for _, p := range posts {
		if int64(p["id"].(float64)) == postID {
			got = p
			break
		}
	}
	if got == nil {
		t.Fatalf("expected post id=%d to be present in list", postID)
	}

	raw, ok := got["categories"]
	if !ok || raw == nil {
		t.Fatalf("expected categories attached, got post=%v", got)
	}

	cats, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected categories array, got %T (%v)", raw, raw)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d (%v)", len(cats), cats)
	}

	want := map[int64]bool{101: true, 102: true}
	for _, c := range cats {
		m, ok := c.(map[string]any)
		if !ok {
			t.Fatalf("expected category object, got %T (%v)", c, c)
		}
		idAny, ok := m["id"]
		if !ok {
			t.Fatalf("category missing id: %v", m)
		}
		id := int64(idAny.(float64))
		if !want[id] {
			t.Fatalf("unexpected category id %d, cats=%v", id, cats)
		}
		delete(want, id)
	}
	if len(want) != 0 {
		t.Fatalf("missing expected categories: %v", want)
	}
}

func TestAPIPostsList_AttachesReactionsCounts(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// Create a known post
	res, err := db.Exec(`
		INSERT INTO posts (author_id, title, body, status, created_at, updated_at)
		VALUES (1, 'ListReacts', 'Body', 'published', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	postID, _ := res.LastInsertId()

	// Create two users for reactions (or use existing seed users if you prefer)
	_, err = db.Exec(`
		INSERT INTO users (username, email, password_hash, is_active, created_at, updated_at)
		VALUES
			('react_u1', 'react1@example.com', 'x', 1, datetime('now'), datetime('now')),
			('react_u2', 'react2@example.com', 'x', 1, datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("seed users: %v", err)
	}

	var u1, u2 int64
	if err := db.QueryRow(`SELECT id FROM users WHERE username='react_u1'`).Scan(&u1); err != nil {
		t.Fatalf("get u1: %v", err)
	}
	if err := db.QueryRow(`SELECT id FROM users WHERE username='react_u2'`).Scan(&u2); err != nil {
		t.Fatalf("get u2: %v", err)
	}

	// Seed reactions: 1 like + 1 dislike
	_, err = db.Exec(`INSERT INTO reactions (user_id, post_id, value, created_at) VALUES (?, ?, ?, datetime('now'))`, u1, postID, 1)
	if err != nil {
		t.Fatalf("seed like: %v", err)
	}
	_, err = db.Exec(`INSERT INTO reactions (user_id, post_id, value, created_at) VALUES (?, ?, ?, datetime('now'))`, u2, postID, -1)
	if err != nil {
		t.Fatalf("seed dislike: %v", err)
	}

	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts?page=1&per_page=50", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, string(body))
	}
	posts := decodePostsList(t, body)

	// Find our post
	var got map[string]any
	for _, p := range posts {
		if int64(p["id"].(float64)) == postID {
			got = p
			break
		}
	}
	if got == nil {
		t.Fatalf("expected post id=%d to be present in list", postID)
	}

	likesAny, ok := got["likes"]
	if !ok {
		t.Fatalf("expected likes field present, post=%v", got)
	}
	dislikesAny, ok := got["dislikes"]
	if !ok {
		t.Fatalf("expected dislikes field present, post=%v", got)
	}

	likes := int(likesAny.(float64))
	dislikes := int(dislikesAny.(float64))

	if likes != 1 || dislikes != 1 {
		t.Fatalf("expected likes=1 dislikes=1, got likes=%d dislikes=%d post=%v", likes, dislikes, got)
	}
}

/*---------
  HELPERS
-----------*/

func seedPosts(t *testing.T, db *sql.DB, authorID int64, n int) {
	t.Helper()

	for i := 0; i < n; i++ {
		_, err := db.Exec(`
			INSERT INTO posts (author_id, title, body, status, created_at, updated_at)
			VALUES (?, ?, ?, 'published', datetime('now', ?), datetime('now', ?))
		`, authorID, "Post "+fmt.Sprint(i), "Body "+fmt.Sprint(i), fmt.Sprintf("-%d seconds", i), fmt.Sprintf("-%d seconds", i))
		if err != nil {
			t.Fatalf("seed posts: %v", err)
		}
	}
}

func decodePostsList(t *testing.T, body []byte) []map[string]any {
	t.Helper()

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("unmarshal posts: %v", err)
	}
	return posts
}

func loginAndGetToken(t *testing.T, h http.Handler, username, password string) string {
	t.Helper()

	loginBody := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d body=%s", rec.Code, rec.Body.String())
	}

	setCookie := rec.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatalf("expected Set-Cookie header on login")
	}
	return strings.Split(strings.Split(setCookie, ";")[0], "=")[1]
}
