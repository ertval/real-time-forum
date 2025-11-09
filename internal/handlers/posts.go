package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type Posts struct{}

func NewPosts() *Posts { return &Posts{} }

// Collection is the main endpoint for posts
func (p *Posts) Collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//lists posts
		page := parseIntOr(r.URL.Query().Get("page"), 1)
		per := parseIntOr(r.URL.Query().Get("per_page"), 20)
		//falls back to 100 so no one can get say 10,000 elements
		if per > 100 {
			per = 100
		}
		if page < 1 {
			page = 1
		}
		//TODO Populate data from database by linking it to API's
		data := []any{}
		meta := map[string]any{"page": page, "per_page": per, "total": 0}
		WriteOK(w, data, meta)

	case http.MethodPost:
		//Creates stub for now
		var in struct {
			Title string `json:"title""`
			Body  string `json:"body"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Title == "" {
			WriteError(w, NewError("BAD_REQUEST", "invalid JSON or missing title", http.StatusBadRequest))
			return
		}
		WriteCreated(w, map[string]any{"id": 1, "title": in.Title, "body": in.Body})

	default:
		WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
	}
}

// Item is the endpoint for a single post and subresources such as /comments, /like
func (p *Posts) Item(w http.ResponseWriter, r *http.Request) {
	tail := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	parts := strings.Split(strings.Trim(tail, "/"), "/")

	//case nothing past tail or empty ID
	if len(parts) == 0 || parts[0] == "" {
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
		return
	}
	postID := parts[0]

	//case /api/v1/posts/{id}
	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			WriteOK(w, map[string]any{"id": postID, "title": "title", "body": "body"}, nil)
		case http.MethodPatch:
			WriteOK(w, map[string]string{"status": "updated"}, nil)
		case http.MethodDelete:
			WriteNoContent(w)
		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		}
		return
	}
	//case /api/v1/posts/{id}/...
	switch parts[1] {
	case "like":
		if r.Method != http.MethodPost {
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
			return
		}
		WriteOK(w, map[string]string{"status": "toggled"}, nil)

	case "comments":
		switch r.Method {
		case http.MethodGet:
			//list comments
			page := parseIntOr(r.URL.Query().Get("page"), 1)
			per := parseIntOr(r.URL.Query().Get("per_page"), 20)
			if per > 100 {
				per = 100
			}
			if page < 1 {
				page = 1
			}
			data := []any{}
			meta := map[string]any{"page": page, "per_page": per, "total": 0, "post_id": postID}
			WriteOK(w, data, meta)
		case http.MethodPost:
			//creates comment stub for now
			var in struct {
				Body string `json:"body"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Body) == "" {
				WriteError(w, NewError("BAD_REQUEST", "body required", http.StatusBadRequest))
				return
			}
			WriteCreated(w, map[string]any{"id": 1, "post_id": postID, "body": in.Body})
		default:
			WriteError(w, NewError("METHOD_NOT_ALLOWED", "method not allowed", http.StatusMethodNotAllowed))
		}

	default:
		WriteError(w, NewError("NOT_FOUND", "route not found", http.StatusNotFound))
	}
}

// DEPRECATED
func (p *Posts) List(w http.ResponseWriter, r *http.Request) {
	//Requests page number & per page elements
	//default page num is 1 and per page elements is 20
	page := parseIntOr(r.URL.Query().Get("page"), 1)
	per := parseIntOr(r.URL.Query().Get("per_page"), 20)
	//falls back to 100 so no one can get say 10,000 elements
	if per > 100 {
		per = 100
	}
	if page < 1 {
		page = 1
	}
	//TODO Populate data from database by linking it to API's
	data := []any{}
	meta := map[string]any{"page": page, "per_page": per, "total": 0}
	WriteOK(w, data, meta)
}

func parseIntOr(s string, def int) int {
	if s == "" {
		return def
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
