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
	Likes       int     `json:"likes"`
	Dislikes    int     `json:"dislikes"`
}

type ListPostsByCategoryParams struct {
	CategoryID int64
	Page       int
	PerPage    int
}

type ListPostsByCategoryResult struct {
	Posts []Post
	Total int
}

type ListPostsParams struct {
	Page    int
	PerPage int
}

type ListPostsResult struct {
	Posts []Post
	Total int
}

// ------------------------------------------------------------
// LIST POSTS
// ------------------------------------------------------------

func ListPosts(ctx context.Context, db *sql.DB, p ListPostsParams) (ListPostsResult, error) {
	// Apply default pagination values if not provided
	p.ApplyDefaults()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	total, err := countPosts(ctx, db)
	if err != nil {
		return ListPostsResult{}, err
	}

	posts, err := fetchPosts(ctx, db, p)
	if err != nil {
		return ListPostsResult{}, err
	}

	if err := attachPostCategories(ctx, db, posts); err != nil {
		return ListPostsResult{}, err
	}

	if err := attachPostReactions(ctx, db, posts); err != nil {
		return ListPostsResult{}, err
	}

	return ListPostsResult{
		Posts: posts,
		Total: total,
	}, nil
}

func (p *ListPostsParams) ApplyDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

func (p *ListPostsByCategoryParams) ApplyDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// ------------------------------------------------------------
// GET POST
// ------------------------------------------------------------

func GetPost(ctx context.Context, db *sql.DB, id int64) (Post, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var post Post

	err := db.QueryRowContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts
		WHERE id = ?
	`, id).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Title,
		&post.Body,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return Post{}, err
	}

	categoryIDs, err := getCategoryIDsByPostID(ctx, db, id)
	if err != nil {
		return Post{}, err
	}
	post.CategoryIDs = categoryIDs

	likes, dislikes, err := CountReactionsForPost(ctx, db, id)
	if err != nil {
		return Post{}, fmt.Errorf("get post reactions: %w", err)
	}
	post.Likes = likes
	post.Dislikes = dislikes

	return post, nil
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

	if err := validateCategories(ctx, db, categoryIDs); err != nil {
		return 0, err
	}

	result, err := db.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now'), datetime('now'))
	`, authorID, title, body)
	if err != nil {
		return 0, fmt.Errorf("create post: %w", err)
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}

	if err := insertPostCategories(ctx, db, postID, categoryIDs); err != nil {
		return 0, err
	}

	return postID, nil
}

// ------------------------------------------------------------
// CREATE HELPERS
// ------------------------------------------------------------

func validateCategories(ctx context.Context, db *sql.DB, categoryIDs []int64) error {
	for _, cid := range categoryIDs {
		var exists bool
		if err := db.QueryRowContext(ctx,
			`SELECT EXISTS(SELECT 1 FROM categories WHERE id = ?)`,
			cid,
		).Scan(&exists); err != nil {
			return fmt.Errorf("validate category: %w", err)
		}
		if !exists {
			return fmt.Errorf("category %d does not exist", cid)
		}
	}
	return nil
}

func insertPostCategories(ctx context.Context, db *sql.DB, postID int64, categoryIDs []int64) error {
	for _, cid := range categoryIDs {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, cid,
		); err != nil {
			return fmt.Errorf("insert category relation: %w", err)
		}
	}
	return nil
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
		return nil
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

	_, _ = db.ExecContext(ctx, `DELETE FROM post_categories WHERE post_id = ?`, id)
	_, err := db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	return err
}

func ListPostsByCategory(
	ctx context.Context,
	db *sql.DB,
	params ListPostsByCategoryParams,
) (ListPostsByCategoryResult, error) {

	params.ApplyDefaults()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Count total posts in category
	var total int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT post_id)
		FROM post_categories
		WHERE category_id = ?
	`, params.CategoryID).Scan(&total); err != nil {
		return ListPostsByCategoryResult{}, err
	}

	offset := (params.Page - 1) * params.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT p.id, p.author_id, p.title, p.body, p.created_at, p.updated_at
		FROM posts p
		JOIN post_categories pc ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, params.CategoryID, params.PerPage, offset)
	if err != nil {
		return ListPostsByCategoryResult{}, err
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
			return ListPostsByCategoryResult{}, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return ListPostsByCategoryResult{}, err
	}

	if err := attachPostCategories(ctx, db, posts); err != nil {
		return ListPostsByCategoryResult{}, err
	}
	if err := attachPostReactions(ctx, db, posts); err != nil {
		return ListPostsByCategoryResult{}, err
	}

	return ListPostsByCategoryResult{
		Posts: posts,
		Total: total,
	}, nil
}

// ============================================================
// LIST POSTS BY AUTHOR (MY POSTS)
// ============================================================

type ListPostsByAuthorParams struct {
	AuthorID int64
	Page     int
	PerPage  int
}

type ListPostsByAuthorResult struct {
	Posts []Post
	Total int
}

func (p *ListPostsByAuthorParams) ApplyDefaults() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 20
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

func ListPostsByAuthor(
	ctx context.Context,
	db *sql.DB,
	params ListPostsByAuthorParams,
) (ListPostsByAuthorResult, error) {

	params.ApplyDefaults()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var total int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM posts
		WHERE author_id = ?
	`, params.AuthorID).Scan(&total); err != nil {
		return ListPostsByAuthorResult{}, err
	}

	offset := (params.Page - 1) * params.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts
		WHERE author_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, params.AuthorID, params.PerPage, offset)
	if err != nil {
		return ListPostsByAuthorResult{}, err
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
			return ListPostsByAuthorResult{}, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return ListPostsByAuthorResult{}, err
	}

	if err := attachPostCategories(ctx, db, posts); err != nil {
		return ListPostsByAuthorResult{}, err
	}
	if err := attachPostReactions(ctx, db, posts); err != nil {
		return ListPostsByAuthorResult{}, err
	}

	return ListPostsByAuthorResult{
		Posts: posts,
		Total: total,
	}, nil
}
