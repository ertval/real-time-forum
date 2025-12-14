package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

//
// ─────────────────────────────────────────────────────────────
//  MODELS
// ─────────────────────────────────────────────────────────────
//

type PublicPost struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	Body       string   `json:"body"`
	Author     string   `json:"author"`
	Categories []string `json:"categories"`
	Likes      int      `json:"likes"`
	Dislikes   int      `json:"dislikes"`
	CreatedAt  string   `json:"created_at"`
}

type ListPublicPostsParams struct {
	Page    int
	PerPage int
	SortBy  string
}

type ListPublicPostsResult struct {
	Posts []PublicPost `json:"posts"`
	Total int          `json:"total"`
}

//
// ─────────────────────────────────────────────────────────────
//  SQL CONSTANTS
// ─────────────────────────────────────────────────────────────
//

// Main public posts query (ORDER BY is injected dynamically)
const sqlListPublicPosts = `
SELECT 
    p.id,
    p.title,
    p.body,
    u.username AS author,
    IFNULL(l.likes, 0) AS likes,
    IFNULL(d.dislikes, 0) AS dislikes,
    p.created_at
FROM posts p
JOIN users u ON u.id = p.author_id

LEFT JOIN (
    SELECT post_id, COUNT(*) AS likes
    FROM reactions
    WHERE value = 1 AND post_id IS NOT NULL
    GROUP BY post_id
) l ON l.post_id = p.id

LEFT JOIN (
    SELECT post_id, COUNT(*) AS dislikes
    FROM reactions
    WHERE value = -1 AND post_id IS NOT NULL
    GROUP BY post_id
) d ON d.post_id = p.id

WHERE p.status = 'published'
ORDER BY %s
LIMIT ? OFFSET ?
`

// Load categories for a specific post
const sqlSelectCategories = `
SELECT c.name
FROM post_categories pc
JOIN categories c ON c.id = pc.category_id
WHERE pc.post_id = ?
`

// Count all published posts
const sqlCountPublishedPosts = `
SELECT COUNT(*) FROM posts WHERE status = 'published'
`

//
// ─────────────────────────────────────────────────────────────
//  MAIN FUNCTION: PUBLIC POSTS LIST
// ─────────────────────────────────────────────────────────────
//

func ListPublicPosts(ctx context.Context, db *sql.DB, p ListPublicPostsParams) (ListPublicPostsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Sorting logic
	order := "datetime(p.created_at) DESC"
	switch strings.ToLower(p.SortBy) {
	case "oldest":
		order = "datetime(p.created_at) ASC"
	case "top":
		order = "likes DESC"
	}

	offset := (p.Page - 1) * p.PerPage

	// Count total published posts
	var total int
	if err := db.QueryRowContext(ctx, sqlCountPublishedPosts).Scan(&total); err != nil {
		return ListPublicPostsResult{}, fmt.Errorf("count posts: %w", err)
	}

	// Final SQL query
	query := fmt.Sprintf(sqlListPublicPosts, order)

	rows, err := db.QueryContext(ctx, query, p.PerPage, offset)
	if err != nil {
		return ListPublicPostsResult{}, fmt.Errorf("query posts: %w", err)
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
			return ListPublicPostsResult{}, fmt.Errorf("scan post: %w", err)
		}

		// Load categories
		if categoryIDs, err := loadCategoriesForPost(ctx, db, post.ID); err == nil {
			post.Categories = categoryIDs
		}

		posts = append(posts, post)
	}

	return ListPublicPostsResult{
		Posts: posts,
		Total: total,
	}, nil
}
