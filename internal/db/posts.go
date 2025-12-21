package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ============================================================
// MODEL
// ============================================================

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

// ============================================================
// LIST POSTS
// ============================================================

type ListPostsParams struct {
	Page    int
	PerPage int
}

type ListPostsResult struct {
	Posts []Post
	Total int
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

func ListPosts(ctx context.Context, db *sql.DB, p ListPostsParams) (ListPostsResult, error) {
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

	return ListPostsResult{Posts: posts, Total: total}, nil
}

// ============================================================
// GET POST
// ============================================================

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

	categories, err := getCategoryIDsByPostID(ctx, db, id)
	if err != nil {
		return Post{}, err
	}
	post.CategoryIDs = categories

	likes, dislikes, err := CountReactionsForPost(ctx, db, id)
	if err != nil {
		return Post{}, err
	}
	post.Likes = likes
	post.Dislikes = dislikes

	return post, nil
}

// ============================================================
// CREATE POST (TRANSACTIONAL)
// ============================================================

func CreatePostWithCategories(
	ctx context.Context,
	db *sql.DB,
	authorID int64,
	title, body string,
	categoryIDs []int64,
) (int64, error) {

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if err := validateCategoriesTx(ctx, tx, categoryIDs); err != nil {
		return 0, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO posts (author_id, title, body, created_at, updated_at)
		VALUES (?, ?, ?, datetime('now'), datetime('now'))
	`, authorID, title, body)
	if err != nil {
		return 0, err
	}

	postID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := insertPostCategoriesTx(ctx, tx, postID, categoryIDs); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return postID, nil
}

// ============================================================
// UPDATE POST
// ============================================================

type UpdatePostInput struct {
	Title *string
	Body  *string
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

	if len(set) == 0 {
		return nil
	}

	set = append(set, "updated_at = datetime('now')")
	args = append(args, id)

	query := `UPDATE posts SET ` + strings.Join(set, ", ") + ` WHERE id = ?`

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// ============================================================
// DELETE POST (TRANSACTIONAL)
// ============================================================

func DeletePost(ctx context.Context, db *sql.DB, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM post_categories WHERE post_id = ?`, id,
	); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx,
		`DELETE FROM posts WHERE id = ?`, id,
	)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return tx.Commit()
}

// ============================================================
// LIST POSTS BY CATEGORY
// ============================================================

type ListPostsByCategoryParams struct {
	CategoryID int64
	Page       int
	PerPage    int
}

type ListPostsByCategoryResult struct {
	Posts []Post
	Total int
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

func ListPostsByCategory(
	ctx context.Context,
	db *sql.DB,
	p ListPostsByCategoryParams,
) (ListPostsByCategoryResult, error) {

	p.ApplyDefaults()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var total int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT post_id)
		FROM post_categories
		WHERE category_id = ?
	`, p.CategoryID).Scan(&total); err != nil {
		return ListPostsByCategoryResult{}, err
	}

	offset := (p.Page - 1) * p.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT p.id, p.author_id, p.title, p.body, p.created_at, p.updated_at
		FROM posts p
		JOIN post_categories pc ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, p.CategoryID, p.PerPage, offset)
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

	if err := attachPostCategories(ctx, db, posts); err != nil {
		return ListPostsByCategoryResult{}, err
	}
	if err := attachPostReactions(ctx, db, posts); err != nil {
		return ListPostsByCategoryResult{}, err
	}

	return ListPostsByCategoryResult{Posts: posts, Total: total}, nil
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
	p ListPostsByAuthorParams,
) (ListPostsByAuthorResult, error) {

	p.ApplyDefaults()

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	var total int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM posts
		WHERE author_id = ?
	`, p.AuthorID).Scan(&total); err != nil {
		return ListPostsByAuthorResult{}, err
	}

	offset := (p.Page - 1) * p.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts
		WHERE author_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, p.AuthorID, p.PerPage, offset)
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

	if err := attachPostCategories(ctx, db, posts); err != nil {
		return ListPostsByAuthorResult{}, err
	}
	if err := attachPostReactions(ctx, db, posts); err != nil {
		return ListPostsByAuthorResult{}, err
	}

	return ListPostsByAuthorResult{Posts: posts, Total: total}, nil
}

// ============================================================
// HELPERS (TX SAFE)
// ============================================================

func validateCategoriesTx(ctx context.Context, tx *sql.Tx, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")
	query := `SELECT COUNT(*) FROM categories WHERE id IN (` + placeholders + `)`

	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	var count int
	if err := tx.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
		return err
	}

	if count != len(ids) {
		return fmt.Errorf("one or more categories do not exist")
	}

	return nil
}

func insertPostCategoriesTx(ctx context.Context, tx *sql.Tx, postID int64, ids []int64) error {
	for _, cid := range ids {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)`,
			postID, cid,
		); err != nil {
			return err
		}
	}
	return nil
}
