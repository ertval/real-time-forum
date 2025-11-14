package db

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"
)

// helper to create an in-memory DB, load schema, and seed minimal data.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Open in-memory SQLite DB
	database, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// Load schema from forum_schema.sql
	schemaPath := "forum_schema.sql"
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		database.Close()
		t.Fatalf("failed to read schema file %s: %v", schemaPath, err)
	}

	if _, err := database.Exec(string(schemaBytes)); err != nil {
		database.Close()
		t.Fatalf("failed to exec schema: %v", err)
	}

	// Seed one user (id=1) and one post (id=1) for comments to reference
	_, err = database.Exec(`
		INSERT INTO users (id, username, email, password_hash, is_active, created_at, updated_at)
		VALUES (1, 'testuser', 'test@example.com', 'hash', 1, datetime('now'), datetime('now'));

		INSERT INTO categories (id, name, slug, created_at)
		VALUES (1, 'Test Category', 'test-category', datetime('now'));

		INSERT INTO posts (id, author_id, title, body, status, category_id, created_at, updated_at)
		VALUES (1, 1, 'Test Post', 'Body of test post', 'published', 1, datetime('now'), datetime('now'));
	`)
	if err != nil {
		database.Close()
		t.Fatalf("failed to seed data: %v", err)
	}

	return database
}

func TestCreateAndGetComment(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a new comment
	input := CreateCommentInput{
		PostID:          1,
		UserID:          1,
		ParentCommentID: nil,
		Body:            "First test comment",
	}

	id, err := CreateComment(ctx, db, input)
	if err != nil {
		t.Fatalf("CreateComment returned error: %v", err)
	}

	// Fetch it back
	comment, err := GetComment(ctx, db, id)
	if err != nil {
		t.Fatalf("GetComment returned error: %v", err)
	}

	if comment.ID != id {
		t.Errorf("expected ID %d, got %d", id, comment.ID)
	}
	if comment.PostID != input.PostID {
		t.Errorf("expected PostID %d, got %d", input.PostID, comment.PostID)
	}
	if comment.UserID != input.UserID {
		t.Errorf("expected UserID %d, got %d", input.UserID, comment.UserID)
	}
	if comment.Body != input.Body {
		t.Errorf("expected Body %q, got %q", input.Body, comment.Body)
	}
	if comment.ParentCommentID != nil {
		t.Errorf("expected ParentCommentID to be nil, got %v", *comment.ParentCommentID)
	}
	if comment.CreatedAt == "" {
		t.Errorf("expected CreatedAt to be set, got empty string")
	}
	if comment.UpdatedAt == "" {
		t.Errorf("expected UpdatedAt to be set, got empty string")
	}
}

func TestListCommentsByPost(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Seed multiple comments for post 1 and some for a different post
	_, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
		VALUES
		  (1, 1, NULL, 'Comment A', datetime('now'), datetime('now')),
		  (1, 1, NULL, 'Comment B', datetime('now'), datetime('now')),
		  (1, 1, NULL, 'Comment C', datetime('now'), datetime('now'));

		INSERT INTO posts (id, author_id, title, body, status, category_id, created_at, updated_at)
		VALUES (2, 1, 'Other Post', 'Other body', 'published', 1, datetime('now'), datetime('now'));

		INSERT INTO comments (post_id, user_id, parent_comment_id, body, created_at, updated_at)
		VALUES (2, 1, NULL, 'Other post comment', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to seed comments: %v", err)
	}

	// List comments for post 1
	res, err := ListCommentsByPost(ctx, db, ListCommentsParams{
		PostID:  1,
		Page:    1,
		PerPage: 10,
	})
	if err != nil {
		t.Fatalf("ListCommentsByPost returned error: %v", err)
	}

	if res.Total != 3 {
		t.Errorf("expected Total 3, got %d", res.Total)
	}
	if len(res.Comments) != 3 {
		t.Errorf("expected 3 comments, got %d", len(res.Comments))
	}

	for _, c := range res.Comments {
		if c.PostID != 1 {
			t.Errorf("expected all comments to have PostID=1, got %d", c.PostID)
		}
		if c.Body == "" {
			t.Errorf("expected non-empty Body")
		}
	}
}
