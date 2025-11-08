package handlers

import (
	"net/http"
	"strconv"
)

type Posts struct{}

func NewPosts() *Posts { return &Posts{} }

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
