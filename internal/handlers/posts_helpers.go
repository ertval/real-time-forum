// Internal/handlers/posts_helpers.go
package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	repository "forum/internal/db"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

const maxUploadSize int64 = 20 << 20

type imageTypeValidationError struct {
	message string
}

func (e *imageTypeValidationError) Error() string {
	return e.message
}

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

func parseMultipartForm(w http.ResponseWriter, r *http.Request) (func(), bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			WriteError(w, r, NewError("PAYLOAD_TOO_LARGE", "upload too large", http.StatusRequestEntityTooLarge))
			return nil, false
		}
		WriteError(w, r, NewError("BAD_REQUEST", "invalid multipart form", http.StatusBadRequest))
		return nil, false
	}

	cleanup := func() {}
	if r.MultipartForm != nil {
		cleanup = func() {
			_ = r.MultipartForm.RemoveAll()
		}
	}
	return cleanup, true
}

func parseImageUpload(w http.ResponseWriter, r *http.Request) (file multipart.File, mime string, hasUpload bool, ok bool) {
	file, fileHeader, err := r.FormFile("image")
	if err == nil {
		mime, err = validateImageType(file, fileHeader.Filename)
		if err != nil {
			_ = file.Close()
			var typeErr *imageTypeValidationError
			if errors.As(err, &typeErr) {
				WriteError(w, r, NewError("BAD_REQUEST", typeErr.Error(), http.StatusBadRequest))
				return nil, "", false, false
			}
			WriteError(w, r, NewError("BAD_REQUEST", "invalid image upload", http.StatusBadRequest))
			return nil, "", false, false
		}
		return file, mime, true, true
	}
	if errors.Is(err, http.ErrMissingFile) {
		return nil, "", false, true
	}

	WriteError(w, r, NewError("BAD_REQUEST", "invalid image upload", http.StatusBadRequest))
	return nil, "", false, false
}

func validateImageType(file io.ReadSeeker, filename string) (string, error) {
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
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			return "", &imageTypeValidationError{
				message: fmt.Sprintf(
					"unsupported image type: detected %s from file content (filename extension is %s). Supported formats: JPEG, PNG, GIF",
					mime,
					ext,
				),
			}
		}
		return "", &imageTypeValidationError{
			message: fmt.Sprintf(
				"unsupported image type: detected %s from file content. Supported formats: JPEG, PNG, GIF",
				mime,
			),
		}
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

// Collects all related Image urls of a post for deletion
func collectPostRelatedImageURLs(ctx context.Context, db *sql.DB, postID int64) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT image_url
		FROM posts
		WHERE id = ? AND image_url IS NOT NULL
		UNION
		SELECT image_url
		FROM comments
		WHERE post_id = ? AND image_url IS NOT NULL
	`, postID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	var urls []string
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		if normalized, ok := normalizeUploadedImageURL(raw); ok {
			if _, exists := seen[normalized]; exists {
				continue
			}
			seen[normalized] = struct{}{}
			urls = append(urls, normalized)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func maybeDeleteUploadedImageByURL(ctx context.Context, db *sql.DB, imageURL string) error {
	normalizedURL, ok := normalizeUploadedImageURL(imageURL)
	if !ok {
		return nil
	}

	var refs int
	if err := db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(1) FROM posts WHERE image_url = ?)
			+ (SELECT COUNT(1) FROM comments WHERE image_url = ?)
	`, normalizedURL, normalizedURL).Scan(&refs); err != nil {
		return err
	}
	if refs > 0 {
		return nil
	}

	diskPath, ok := uploadedImageDiskPath(normalizedURL)
	if !ok {
		return nil
	}
	if err := os.Remove(diskPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func normalizeUploadedImageURL(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", false
	}

	if cut := strings.IndexAny(value, "?#"); cut >= 0 {
		value = value[:cut]
	}

	const prefix = "/static/uploads/"
	if !strings.HasPrefix(value, prefix) {
		return "", false
	}

	filename := strings.TrimPrefix(value, prefix)
	if filename == "" {
		return "", false
	}
	if filename != filepath.Base(filename) {
		return "", false
	}

	return prefix + filename, true
}

func uploadedImageDiskPath(imageURL string) (string, bool) {
	normalizedURL, ok := normalizeUploadedImageURL(imageURL)
	if !ok {
		return "", false
	}
	filename := strings.TrimPrefix(normalizedURL, "/static/uploads/")
	return filepath.Join("web", "static", "uploads", filename), true
}
