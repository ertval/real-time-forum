// internal/tests/categories_test.go
package tests

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
)

// ============================================================
// DTOs
// ============================================================

type categoryDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ============================================================
// Helpers
// ============================================================

func decodeEnvelope(t *testing.T, body []byte) apiEnvelope {
	t.Helper()

	var env apiEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("failed to decode envelope: %v", err)
	}

	if env.Error != nil {
		t.Fatalf("unexpected API error: %+v", env.Error)
	}

	return env
}

// ============================================================
// Tests
// ============================================================

func TestCategoriesList(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(t, h, http.MethodGet, "/api/v1/categories", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	env := decodeEnvelope(t, rec.Body.Bytes())

	var categories []categoryDTO
	if err := json.Unmarshal(env.Data, &categories); err != nil {
		t.Fatalf("failed to decode categories: %v", err)
	}

	if len(categories) == 0 {
		t.Fatalf("expected at least one category")
	}
}

// ------------------------------------------------------------

func TestCategoriesCreateAndGet(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(
		t,
		h,
		http.MethodPost,
		"/api/v1/categories",
		[]byte(`{"name":"New Category"}`),
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	env := decodeEnvelope(t, rec.Body.Bytes())

	var created categoryDTO
	if err := json.Unmarshal(env.Data, &created); err != nil {
		t.Fatalf("failed to decode created category: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}

	rec2 := doReq(
		t,
		h,
		http.MethodGet,
		"/api/v1/categories/"+strconv.FormatInt(created.ID, 10),
		nil,
	)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	env2 := decodeEnvelope(t, rec2.Body.Bytes())

	var fetched categoryDTO
	json.Unmarshal(env2.Data, &fetched)

	if fetched.Name != "New Category" {
		t.Fatalf("expected name 'New Category', got %q", fetched.Name)
	}
}

// ------------------------------------------------------------

func TestCategoriesUpdate(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(
		t,
		h,
		http.MethodPost,
		"/api/v1/categories",
		[]byte(`{"name":"Old"}`),
	)

	env := decodeEnvelope(t, rec.Body.Bytes())

	var cat categoryDTO
	json.Unmarshal(env.Data, &cat)

	rec2 := doReq(
		t,
		h,
		http.MethodPatch,
		"/api/v1/categories/"+strconv.FormatInt(cat.ID, 10),
		[]byte(`{"name":"Updated Category"}`),
	)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}

	env2 := decodeEnvelope(t, rec2.Body.Bytes())

	var updated categoryDTO
	json.Unmarshal(env2.Data, &updated)

	if updated.Name != "Updated Category" {
		t.Fatalf("expected updated name, got %q", updated.Name)
	}
}

// ------------------------------------------------------------

func TestCategoriesDelete(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(
		t,
		h,
		http.MethodPost,
		"/api/v1/categories",
		[]byte(`{"name":"To Delete"}`),
	)

	env := decodeEnvelope(t, rec.Body.Bytes())

	var cat categoryDTO
	json.Unmarshal(env.Data, &cat)

	rec2 := doReq(
		t,
		h,
		http.MethodDelete,
		"/api/v1/categories/"+strconv.FormatInt(cat.ID, 10),
		nil,
	)

	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec2.Code)
	}

	rec3 := doReq(
		t,
		h,
		http.MethodGet,
		"/api/v1/categories/"+strconv.FormatInt(cat.ID, 10),
		nil,
	)

	if rec3.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec3.Code)
	}
}
