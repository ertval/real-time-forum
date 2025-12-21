package tests

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	repository "forum/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// SQLite FK enforcement is OFF by default.
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		t.Fatalf("enable foreign_keys: %v", err)
	}

	schemaBytes, err := os.ReadFile("../db/forum_schema.sql")
	if err != nil {
		t.Fatalf("read schema: %v", err)
	}

	// Execute schema.
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		t.Fatalf("exec schema: %v", err)
	}

	return db
}

func seedUserAndPost(t *testing.T, db *sql.DB) (userID, postID int64) {
	t.Helper()

	res, err := db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`, "user1", "u1@example.com", "hash")
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, _ = res.LastInsertId()

	res, err = db.Exec(`INSERT INTO posts (author_id, title, body, status) VALUES (?, ?, ?, ?)`, userID, "t", "b", "published")
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}
	postID, _ = res.LastInsertId()

	return userID, postID
}

func TestCreateGetUpdateDeleteComment(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	userID, postID := seedUserAndPost(t, db)

	ctx := context.Background()

	// Create
	commentID, err := repository.CreateComment(ctx, db, repository.CreateCommentInput{
		PostID: postID,
		UserID: userID,
		Body:   "hello",
	})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}
	if commentID == 0 {
		t.Fatalf("expected non-zero commentID")
	}

	// Get
	c, err := repository.GetComment(ctx, db, commentID)
	if err != nil {
		t.Fatalf("GetComment: %v", err)
	}
	if c.ID != commentID || c.PostID != postID || c.UserID != userID {
		t.Fatalf("unexpected comment: %+v", c)
	}
	if c.Body != "hello" {
		t.Fatalf("expected body 'hello', got %q", c.Body)
	}
	if c.ParentCommentID != nil {
		t.Fatalf("expected nil parent_comment_id")
	}
	if c.Likes != 0 || c.Dislikes != 0 {
		t.Fatalf("expected 0 likes/dislikes, got %d/%d", c.Likes, c.Dislikes)
	}
	if strings.TrimSpace(c.CreatedAt) == "" {
		t.Fatalf("expected CreatedAt to be set")
	}

	// Update
	newBody := "updated"
	if err := repository.UpdateComment(ctx, db, commentID, repository.UpdateCommentInput{Body: &newBody}); err != nil {
		t.Fatalf("UpdateComment: %v", err)
	}
	c2, err := repository.GetComment(ctx, db, commentID)
	if err != nil {
		t.Fatalf("GetComment after update: %v", err)
	}
	if c2.Body != "updated" {
		t.Fatalf("expected updated body, got %q", c2.Body)
	}

	// Delete
	if err := repository.DeleteComment(ctx, db, commentID); err != nil {
		t.Fatalf("DeleteComment: %v", err)
	}
	_, err = repository.GetComment(ctx, db, commentID)
	if err == nil {
		t.Fatalf("expected error after delete")
	}
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("expected sql.ErrNoRows after delete, got %v", err)
	}
}

func TestUpdateComment_NoFieldsIsNoop(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	userID, postID := seedUserAndPost(t, db)

	ctx := context.Background()
	commentID, err := repository.CreateComment(ctx, db, repository.CreateCommentInput{PostID: postID, UserID: userID, Body: "x"})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	// Should not error and should not change updated_at in a way we can easily assert,
	// but it must at least return nil.
	if err := repository.UpdateComment(ctx, db, commentID, repository.UpdateCommentInput{}); err != nil {
		t.Fatalf("UpdateComment noop: %v", err)
	}
}

func TestCreateComment_FKEnforced(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	ctx := context.Background()

	// Non-existent post_id/user_id should fail under FK enforcement.
	_, err := repository.CreateComment(ctx, db, repository.CreateCommentInput{PostID: 999, UserID: 999, Body: "x"})
	if err == nil {
		t.Fatalf("expected fk error")
	}
}

func TestCommentReactions_CountsFlowThroughGetComment(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	userID, postID := seedUserAndPost(t, db)
	// Seed a second user for another reaction.
	res, err := db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`, "user2", "u2@example.com", "hash")
	if err != nil {
		t.Fatalf("insert user2: %v", err)
	}
	user2, _ := res.LastInsertId()

	ctx := context.Background()
	commentID, err := repository.CreateComment(ctx, db, repository.CreateCommentInput{PostID: postID, UserID: userID, Body: "hello"})
	if err != nil {
		t.Fatalf("CreateComment: %v", err)
	}

	// Like by user1, dislike by user2.
	_, err = db.Exec(`INSERT INTO reactions (user_id, comment_id, value) VALUES (?, ?, ?)`, userID, commentID, 1)
	if err != nil {
		t.Fatalf("insert like reaction: %v", err)
	}
	_, err = db.Exec(`INSERT INTO reactions (user_id, comment_id, value) VALUES (?, ?, ?)`, user2, commentID, -1)
	if err != nil {
		t.Fatalf("insert dislike reaction: %v", err)
	}

	// Give SQLite a tick if your reaction counting uses time-based constraints (it shouldn't).
	time.Sleep(5 * time.Millisecond)

	c, err := repository.GetComment(ctx, db, commentID)
	if err != nil {
		t.Fatalf("GetComment: %v", err)
	}
	if c.Likes != 1 || c.Dislikes != 1 {
		t.Fatalf("expected likes/dislikes 1/1, got %d/%d", c.Likes, c.Dislikes)
	}
}
