// Internal/handlers/posts_helpers.go
package handlers

import (
	"database/sql"
	"errors"
	repository "forum/internal/db"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func resolvePostRoute(w http.ResponseWriter, r *http.Request) (postID int64, action string, ok bool) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 0 || parts[0] == "" {
		notFound(w, r)
		return 0, "", false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		WriteError(w, r, NewError("BAD_REQUEST", "invalid post id", http.StatusBadRequest))
		return 0, "", false
	}

	if len(parts) > 1 {
		action = strings.Split(parts[1], "?")[0]
	}

	return id, action, true
}

func parsePostStatusFilter(r *http.Request) (*string, error) {
	status := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status")))

	switch status {
	case "", "all":
		return nil, nil
	case "draft", "published":
		return &status, nil
	default:
		return nil, errors.New("invalid status filter")
	}
}

func (p *PostsHandler) requirePostAuthor(w http.ResponseWriter, r *http.Request, postID int64, userID int64) bool {
	authorID, err := repository.GetPostAuthorID(r.Context(), p.conn, postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r)
			return false
		}
		log.Printf("failed to load post author: %v", err)
		WriteError(w, r, NewError(
			"INTERNAL_SERVER_ERROR",
			"error loading post",
			http.StatusInternalServerError,
		))
		return false
	}

	if authorID != userID {
		WriteError(w, r, NewError(
			"FORBIDDEN",
			"not allowed",
			http.StatusForbidden,
		))
		return false
	}

	return true
}

func parseCategoryIDs(values []string) ([]int64, error) {
	if len(values) == 0 {
		return nil, nil
	}
	ids := make([]int64, 0, len(values))
	for _, raw := range values {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil, errors.New("invalid category_id")
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			return nil, errors.New("invalid category_id")
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func validateImageType(file io.ReadSeeker) (string, error) {
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	mime := http.DetectContentType(buf[:n])
	switch mime {
	case "image/jpeg", "image/png", "image/gif":
		return mime, nil
	default:
		return "", errors.New("unsupported image type")
	}
}

func saveUploadedImage(file io.Reader, mime string) (string, string, error) {
	var ext string
	switch mime {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	default:
		return "", "", errors.New("unsupported image type")
	}

	dir := filepath.Join("web", "static", "uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", "", err
	}

	filename := uuid.New().String() + ext
	diskPath := filepath.Join(dir, filename)

	dst, err := os.OpenFile(diskPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		_ = os.Remove(diskPath)
		return "", "", err
	}

	return "/static/uploads/" + filename, diskPath, nil
}
