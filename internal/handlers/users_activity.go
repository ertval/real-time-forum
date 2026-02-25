// internal/handlers/users_activity.go
package handlers

import (
	"log"
	"net/http"

	repository "forum/internal/db"
	"forum/internal/middleware"
)

type postsActivitySection struct {
	Items      []repository.Post `json:"items"`
	Pagination *PaginationMeta   `json:"pagination"`
}

type commentsActivitySection struct {
	Items      []repository.UserActivityComment `json:"items"`
	Pagination *PaginationMeta                  `json:"pagination"`
}

type userActivityResponse struct {
	CreatedPosts  postsActivitySection    `json:"created_posts"`
	LikedPosts    postsActivitySection    `json:"liked_posts"`
	DislikedPosts postsActivitySection    `json:"disliked_posts"`
	Comments      commentsActivitySection `json:"comments"`
}

func (u *UsersHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		WriteError(w, r, NewError(
			"UNAUTHORIZED",
			"login required",
			http.StatusUnauthorized,
		))
		return
	}

	page, perPage := sanitizePagination(r)

	statusPtr, err := parsePostStatusFilter(r)
	if err != nil {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid status", http.StatusBadRequest))
		return
	}

	created, err := repository.ListPostsByAuthor(
		r.Context(),
		u.conn,
		repository.ListPostsByAuthorParams{
			AuthorID: userID,
			Page:     page,
			PerPage:  perPage,
			Status:   statusPtr,
		},
	)
	if err != nil {
		log.Printf("failed to list created posts for activity: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading user activity", http.StatusInternalServerError))
		return
	}

	liked, err := repository.ListPostsByUserReaction(
		r.Context(),
		u.conn,
		repository.ListPostsByUserReactionParams{
			UserID:   userID,
			Page:     page,
			PerPage:  perPage,
			Reaction: repository.ReactionLike,
		},
	)
	if err != nil {
		log.Printf("failed to list liked posts for activity: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading user activity", http.StatusInternalServerError))
		return
	}

	disliked, err := repository.ListPostsByUserReaction(
		r.Context(),
		u.conn,
		repository.ListPostsByUserReactionParams{
			UserID:   userID,
			Page:     page,
			PerPage:  perPage,
			Reaction: repository.ReactionDislike,
		},
	)
	if err != nil {
		log.Printf("failed to list disliked posts for activity: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading user activity", http.StatusInternalServerError))
		return
	}

	comments, err := repository.ListUserCommentsWithPost(
		r.Context(),
		u.conn,
		repository.ListUserCommentsWithPostParams{
			UserID:  userID,
			Page:    page,
			PerPage: perPage,
		},
	)
	if err != nil {
		log.Printf("failed to list comments for activity: %v", err)
		WriteError(w, r, NewError("INTERNAL_SERVER_ERROR", "error loading user activity", http.StatusInternalServerError))
		return
	}

	resp := userActivityResponse{
		CreatedPosts: postsActivitySection{
			Items:      created.Posts,
			Pagination: makePaginationMeta(page, perPage, created.Total),
		},
		LikedPosts: postsActivitySection{
			Items:      liked.Posts,
			Pagination: makePaginationMeta(page, perPage, liked.Total),
		},
		DislikedPosts: postsActivitySection{
			Items:      disliked.Posts,
			Pagination: makePaginationMeta(page, perPage, disliked.Total),
		},
		Comments: commentsActivitySection{
			Items:      comments.Comments,
			Pagination: makePaginationMeta(page, perPage, comments.Total),
		},
	}

	WriteOK(w, resp, nil)
}

func makePaginationMeta(page, perPage, total int) *PaginationMeta {
	return &PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	}
}
