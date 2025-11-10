package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID         int64  `json:"id"`
	AuthorID   int64  `json:"author_id"`
	Title      string `json:"title"`
	Body       string `json:"body"`
	CategoryID *int64 `json:"category_id,omitempty"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

// ListPosts returns paginated posts with optional filters.

type ListPostsParams struct {
	Page    int
	PerPage int
}

type ListPostsResult struct {
	Posts []Post
	Total int
}

// ListPosts returns posts with pagination,newest first.
// No text search, and no filters yet
// TODO implement filters and text search
func ListPosts(ctx context.Context, db *sql.DB, p ListPostsParams) (ListPostsResult, error) {
	// sane defaults
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
	offset := (p.Page - 1) * p.PerPage

	// 1) total count (no WHERE at all)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var total int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&total); err != nil {
		return ListPostsResult{}, fmt.Errorf("count posts: %w", err)
	}

	// 2) page of rows
	rows, err := db.QueryContext(ctx, `
		SELECT id, author_id, title, body, category_id, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, p.PerPage, offset)
	if err != nil {
		fmt.Printf("ListPosts SELECT error: %v\n", err) // <-- add this line
		return ListPostsResult{}, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var out []Post
	for rows.Next() {
		var p Post
		//special types (Scan would normally fail a col is NULL
		//and a regular type is used
		var cat sql.NullInt64
		if err := rows.Scan(&p.ID, &p.AuthorID, &p.Title, &p.Body, &cat, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return ListPostsResult{}, fmt.Errorf("scan post: %w", err)
		}
		//category ID
		if cat.Valid {
			v := cat.Int64
			p.CategoryID = &v
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return ListPostsResult{}, fmt.Errorf("rows posts: %w", err)
	}

	return ListPostsResult{Posts: out, Total: total}, nil
}

func GetPost(ctx context.Context, db *sql.DB, id int64) (Post, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	q := `SELECT id, author_id, title, body, category_id, created_at, updated_at FROM posts WHERE id = ?`
	var p Post
	//Post category
	var cat sql.NullInt64
	//Updated at
	if err := db.QueryRowContext(ctx, q, id).Scan(&p.ID, &p.AuthorID, &p.Title, &p.Body, &cat, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return Post{}, err
		}
		return Post{}, fmt.Errorf("get post: %w", err)
	}
	if cat.Valid {
		v := cat.Int64
		p.CategoryID = &v
	}
	return p, nil
}

type CreatePostInput struct {
	AuthorID   int64
	Title      string
	Body       string
	CategoryID *int64
}

// TODO link DB
func CreatePost(ctx context.Context, db *sql.DB, in CreatePostInput) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	q := `INSERT INTO posts (author_id, title, body, category_id, created_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)`
	res, err := db.ExecContext(ctx, q, in.AuthorID, in.Title, in.Body, in.CategoryID)
	if err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return id, nil
}

type UpdatePostInput struct {
	Title      *string
	Body       *string
	CategoryID *int64
}

func UpdatePost(ctx context.Context, db *sql.DB, id int64, in UpdatePostInput) error {
	set := []string{}
	args := []any{}

	if in.Title != nil {
		set = append(set, "title = ?")
		args = append(args, *in.Title)
	}
	if in.Body != nil {
		set = append(set, "body = ?")
		args = append(args, *in.Body)
	}
	if in.CategoryID != nil {
		set = append(set, "category_id = ?")
		args = append(args, *in.CategoryID)
	}
	//means there's nothing to update
	if len(set) == 0 {
		return nil
	}
	set = append(set, "updated_at = CURRENT_TIMESTAMP")
	q := `UPDATE posts SET ` + strings.Join(set, ", ") + ` WHERE id = ?`
	args = append(args, id)

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	return nil
}

func DeletePost(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err := db.ExecContext(ctx, "DELETE FROM posts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	return nil
}
