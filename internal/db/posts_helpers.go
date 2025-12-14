package db

import (
	"context"
	"database/sql"
	"fmt"
)

func countPosts(ctx context.Context, db *sql.DB) (int, error) {
	var total int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&total); err != nil {
		return 0, fmt.Errorf("count posts: %w", err)
	}
	return total, nil
}

func fetchPosts(ctx context.Context, db *sql.DB, p ListPostsParams) ([]Post, error) {
	offset := (p.Page - 1) * p.PerPage

	rows, err := db.QueryContext(ctx, `
		SELECT id, author_id, title, body, created_at, updated_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, p.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
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
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func attachPostCategories(ctx context.Context, db *sql.DB, posts []Post) error {
	for i := range posts {
		categoryIDs, err := getCategoryIDsByPostID(ctx, db, posts[i].ID)
		if err != nil {
			return err
		}
		posts[i].CategoryIDs = categoryIDs
	}
	return nil
}

func getCategoryIDsByPostID(ctx context.Context, db *sql.DB, postID int64) ([]int64, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT category_id FROM post_categories WHERE post_id = ?`,
		postID,
	)
	if err != nil {
		return nil, fmt.Errorf("load categories: %w", err)
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		ids = append(ids, cid)
	}
	return ids, nil
}
