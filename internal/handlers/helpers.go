package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

func sanitizePagination(r *http.Request) (page int, per int) {
	page = parseIntOr(r.URL.Query().Get("page"), 1)
	per = parseIntOr(r.URL.Query().Get("per_page"), 20)

	if page < 1 {
		page = 1
	}
	if per < 1 {
		per = 20
	}
	if per > 100 {
		per = 100
	}
	return
}

func sanitizeSort(r *http.Request) string {
	sort := strings.ToLower(r.URL.Query().Get("sort"))
	if sort == "" {
		return "newest"
	}
	return sort
}

func buildMeta(page, per, total int, extra map[string]any) map[string]any {
	meta := map[string]any{
		"page":     page,
		"per_page": per,
		"total":    total,
	}
	for k, v := range extra {
		meta[k] = v
	}
	return meta
}

func parseIntOr(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return def
}
