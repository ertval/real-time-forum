package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestPublicPostsList(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	// -------------------------
	// Seed user with unique username + email
	// -------------------------
	username := fmt.Sprintf("user_%d", time.Now().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)

	res, err := db.Exec(`
        INSERT INTO users (username, email, password_hash)
        VALUES (?, ?, 'x');
    `, username, email)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}
	userID, _ := res.LastInsertId()

	// -------------------------
	// Seed post
	// -------------------------
	res, err = db.Exec(`
        INSERT INTO posts (author_id, title, body, status)
        VALUES (?, 'Seed Post', 'Seed post body', 'published');
    `, userID)
	if err != nil {
		t.Fatalf("failed to insert post: %v", err)
	}
	postID, _ := res.LastInsertId()

	// -------------------------
	// Seed categories
	// -------------------------
	_, _ = db.Exec(`
        INSERT INTO categories (name, slug)
        VALUES ('Tech', 'tech'), ('Fun', 'fun');
    `)

	// Attach categories to post
	_, _ = db.Exec(`
        INSERT INTO post_categories (post_id, category_id)
        VALUES (?, 1), (?, 2);
    `, postID, postID)

	// -------------------------
	// Seed reactions (1 like, 1 dislike)
	// -------------------------
	_, _ = db.Exec(`INSERT INTO reactions (user_id, post_id, value) VALUES (?, ?, 1);`, userID, postID)
	_, _ = db.Exec(`INSERT INTO reactions (user_id, post_id, value) VALUES (999, ?, -1);`, postID)

	// -------------------------
	// Perform API request
	// -------------------------
	w, body := doRequest(t, h, http.MethodGet, "/api/v1/posts/public?page=1&per_page=5", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, string(body))
	}

	// Decode envelope
	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to unmarshal envelope: %v\nraw: %s", err, string(body))
	}

	// Decode posts array
	var posts []map[string]any
	if err := json.Unmarshal(env.Data, &posts); err != nil {
		t.Fatalf("failed to decode posts array: %v\nraw: %s", err, string(env.Data))
	}

	if len(posts) == 0 {
		t.Fatalf("expected at least 1 post, got 0\nraw: %s", string(env.Data))
	}

	// -------------------------
	// Find the correct post (the one created in THIS test)
	// -------------------------
	var found map[string]any
	for _, p := range posts {
		if p["author"] == username {
			found = p
			break
		}
	}

	if found == nil {
		t.Fatalf("expected to find post from user %s\nraw: %s", username, string(env.Data))
	}

	// -------------------------
	// Validations
	// -------------------------

	if found["likes"].(float64) != 1 {
		t.Errorf("expected 1 like, got %v", found["likes"])
	}

	if found["dislikes"].(float64) != 1 {
		t.Errorf("expected 1 dislike, got %v", found["dislikes"])
	}

	if found["author"] != username {
		t.Errorf("expected author '%s', got %v", username, found["author"])
	}
}
