package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Post struct {
	ID          int64   `json:"id"`
	AuthorID    int64   `json:"author_id"`
	Title       string  `json:"title"`
	Body        string  `json:"body"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
	CategoryIDs []int64 `json:"category_ids,omitempty"`
}

// ------------------------------------------------------------
// LIST POSTS
// ------------------------------------------------------------

type ListPostsParams struct {
	Page    int
	PerPage int
}

type ListPostsResult struct {
	Posts []Post
	Total int
}

func ListPosts(ctx context.Context, db *sql.DB, p ListPostsParams) (ListPostsResult, error) {
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

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// total count
	var total int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&total); err != nil {
		return ListPostsResult{}, fmt.Errorf("count posts: %w", err)
	}

	// fetch posts
	rows, err := db.QueryContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, p.PerPage, offset)
	if err != nil {
		return ListPostsResult{}, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	var posts []Post

	for rows.Next() {
		var post Post

		if err := rows.Scan(
			&post.ID,
			&post.AuthorID,
			&post.Title,
			&post.Body,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return ListPostsResult{}, fmt.Errorf("scan post: %w", err)
		}

		// Load categories
		categoryRows, err := db.QueryContext(ctx,
			`SELECT category_id FROM post_categories WHERE post_id = ?`,
			post.ID,
		)
		if err == nil {
			var catIDs []int64
			for categoryRows.Next() {
				var cid int64
				if err := categoryRows.Scan(&cid); err == nil {
					catIDs = append(catIDs, cid)
				}
			}
			categoryRows.Close()
			post.CategoryIDs = catIDs
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return ListPostsResult{}, fmt.Errorf("rows posts: %w", err)
	}

	return ListPostsResult{Posts: posts, Total: total}, nil
}

// ------------------------------------------------------------
// GET POST
// ------------------------------------------------------------

func GetPost(ctx context.Context, db *sql.DB, id int64) (Post, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var p Post

	err := db.QueryRowContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts WHERE id = ?
	`, id).Scan(
		&p.ID,
		&p.AuthorID,
		&p.Title,
		&p.Body,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return Post{}, err
	}

	rows, err := db.QueryContext(ctx,
		`SELECT category_id FROM post_categories WHERE post_id = ?`,
		id,
	)
	if err == nil {
		var cats []int64
		for rows.Next() {
			var cid int64
			if err := rows.Scan(&cid); err == nil {
				cats = append(cats, cid)
			}
		}
		rows.Close()
		p.CategoryIDs = cats
	}

	return p, nil
}

// ------------------------------------------------------------
// CREATE POST WITH CATEGORIES
// ------------------------------------------------------------

func CreatePostWithCategories(
	ctx context.Context,
	db *sql.DB,
	authorID int64,
	title, body string,
	categoryIDs []int64,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Validate categories exist
	for _, cid := range categoryIDs {
		var exists bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)`,
			cid,
		).Scan(&exists); err != nil {
			return 0, fmt.Errorf("validate category: %w", err)
		}
		if !exists {
			return 0, fmt.Errorf("category %d does not exist", cid)
		}
	}

	// Create post
	res, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now'), datetime('now'))
	`, authorID, title, body)
	if err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	// Insert categories
	for _, cid := range categoryIDs {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, cid,
		); err != nil {
			return 0, fmt.Errorf("insert category relation: %w", err)
		}
	}

	return postID, nil
}

// ------------------------------------------------------------
// UPDATE POST
// ------------------------------------------------------------

type UpdatePostInput struct {
	Title *string
	Body  *string
}

func UpdatePost(ctx context.Context, db *sql.DB, id int64, in UpdatePostInput) error {
	setParts := []string{}
	args := []any{}

	if in.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *in.Title)
	}

	if in.Body != nil {
		setParts = append(setParts, "body = ?")
		args = append(args, *in.Body)
	}

	if len(setParts) == 0 {
		return nil // nothing to update
	}

	setParts = append(setParts, "updated_at = datetime('now')")
	args = append(args, id)

	query := `UPDATE posts SET ` + strings.Join(setParts, ", ") + ` WHERE id = ?`

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	_, err := db.ExecContext(ctx, query, args...)
	return err
}

// ------------------------------------------------------------
// DELETE POST
// ------------------------------------------------------------

func DeletePost(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// remove category relations
	_, _ = db.ExecContext(ctx, `DELETE FROM post_categories WHERE post_id = ?`, id)

	// delete post
	_, err := db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}
