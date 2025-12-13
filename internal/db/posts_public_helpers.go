package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

//
// ------------------------------------------------------------
// PUBLIC POSTS HELPERS (PRIVATE)
// ------------------------------------------------------------
// These helpers are intentionally unexported.
// They support ListPublicPosts orchestration only.
// ------------------------------------------------------------

// Resolve ORDER BY clause for public posts (safe, controlled)
func resolvePublicPostOrder(sortBy string) string {
	switch strings.ToLower(sortBy) {
	case "oldest":
		return "datetime(p.created_at) ASC"
	case "top":
		return "likes DESC"
	default:
		return "datetime(p.created_at) DESC"
	}
}

// Count all published posts
func countPublishedPosts(ctx context.Context, db *sql.DB) (int, error) {
	var total int
	if err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM posts WHERE status = 'published'`,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count published posts: %w", err)
	}
	return total, nil
}

// Fetch public posts WITHOUT categories
func fetchPublicPosts(
	ctx context.Context,
	db *sql.DB,
	p ListPublicPostsParams,
) ([]PublicPost, error) {

	order := resolvePublicPostOrder(p.SortBy)
	offset := (p.Page - 1) * p.PerPage

	query := fmt.Sprintf(sqlListPublicPosts, order)

	rows, err := db.QueryContext(ctx, query, p.PerPage, offset)
	if err != nil {
		return nil, fmt.Errorf("query public posts: %w", err)
	}
	defer rows.Close()

	var posts []PublicPost

	for rows.Next() {
		var post PublicPost
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Body,
			&post.Author,
			&post.Likes,
			&post.Dislikes,
			&post.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan public post: %w", err)
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows public posts: %w", err)
	}

	return posts, nil
}

// Attach category names to each public post (N+1 by design, easy to batch later)
func attachPublicPostCategories(
	ctx context.Context,
	db *sql.DB,
	posts []PublicPost,
) error {

	for i := range posts {
		cats, err := loadCategoriesForPost(ctx, db, posts[i].ID)
		if err != nil {
			return fmt.Errorf("load categories for post %d: %w", posts[i].ID, err)
		}
		posts[i].Categories = cats
	}
	return nil
}

// Load category names for a single post
func loadCategoriesForPost(
	ctx context.Context,
	db *sql.DB,
	postID int64,
) ([]string, error) {

	rows, err := db.QueryContext(ctx, sqlSelectCategories, postID)
	if err != nil {
		return nil, fmt.Errorf("select categories: %w", err)
	}
	defer rows.Close()

	var cats []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		cats = append(cats, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows categories: %w", err)
	}

	return cats, nil
}
