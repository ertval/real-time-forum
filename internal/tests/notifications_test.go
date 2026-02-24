// internal/tests/notifications_test.go
package tests

import (
	"context"
	"database/sql"
	"testing"

	repository "forum/internal/db"
)

/*-------------------------------------------------
  Helpers (local to this test file)
--------------------------------------------------*/

func createTestUser(t *testing.T, db *sql.DB, username string) int64 {
	t.Helper()

	res, err := db.Exec(`
		INSERT INTO users (username, email, password_hash)
		VALUES (?, ?, ?)
	`, username, username+"@test.com", "hash")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	return id
}

func createTestPost(t *testing.T, db *sql.DB, authorID int64) int64 {
	t.Helper()

	res, err := db.Exec(`
		INSERT INTO posts (author_id, title, body)
		VALUES (?, 'title', 'body')
	`, authorID)
	if err != nil {
		t.Fatalf("create post: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id: %v", err)
	}

	return id
}

func countNotifications(t *testing.T, db *sql.DB) int {
	t.Helper()

	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM notifications`).Scan(&count)
	if err != nil {
		t.Fatalf("count notifications: %v", err)
	}

	return count
}

/*-------------------------------------------------
  Tests
--------------------------------------------------*/

func TestNotification_OnPostLike(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	authorID := createTestUser(t, db, "author")
	userID := createTestUser(t, db, "user")
	postID := createTestPost(t, db, authorID)

	_, err := repository.ToggleReaction(ctx, db, userID, postID, 1, "post")
	if err != nil {
		t.Fatalf("toggle reaction: %v", err)
	}

	if count := countNotifications(t, db); count != 1 {
		t.Fatalf("expected 1 notification, got %d", count)
	}
}

func TestNotification_NoDuplicateOnRepeatedLike(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	authorID := createTestUser(t, db, "author")
	userID := createTestUser(t, db, "user")
	postID := createTestPost(t, db, authorID)

	// like
	_, _ = repository.ToggleReaction(ctx, db, userID, postID, 1, "post")
	// remove
	_, _ = repository.ToggleReaction(ctx, db, userID, postID, 1, "post")
	// like again
	_, _ = repository.ToggleReaction(ctx, db, userID, postID, 1, "post")

	if count := countNotifications(t, db); count != 1 {
		t.Fatalf("expected 1 notification after repeated likes, got %d", count)
	}
}

func TestNotification_NoSelfNotification(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	authorID := createTestUser(t, db, "author")
	postID := createTestPost(t, db, authorID)

	_, _ = repository.ToggleReaction(ctx, db, authorID, postID, 1, "post")

	if count := countNotifications(t, db); count != 0 {
		t.Fatalf("expected 0 self-notifications, got %d", count)
	}
}

func TestNotification_OnComment(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	authorID := createTestUser(t, db, "author")
	userID := createTestUser(t, db, "user")
	postID := createTestPost(t, db, authorID)

	_, err := repository.CreateComment(ctx, db, repository.CreateCommentInput{
		PostID: postID,
		UserID: userID,
		Body:   "hello",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}

	if count := countNotifications(t, db); count != 1 {
		t.Fatalf("expected 1 notification for comment, got %d", count)
	}
}
