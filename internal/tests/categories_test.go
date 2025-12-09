package tests

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestCategoriesList(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(t, h, "GET", "/api/v1/categories", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var env apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env)

	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var cats []map[string]any
	json.Unmarshal(env.Data, &cats)

	if len(cats) == 0 {
		t.Fatalf("expected at least 1 category")
	}
}

func TestCategoriesCreate(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	body := `{"name":"New Category"}`
	rec := doReq(t, h, "POST", "/api/v1/categories", []byte(body))

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d, body=%s", rec.Code, rec.Body.String())
	}

	var env apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env)

	if env.Error != nil {
		t.Fatalf("unexpected error: %+v", env.Error)
	}

	var cat map[string]any
	json.Unmarshal(env.Data, &cat)

	if cat["name"] != "New Category" {
		t.Errorf("expected name 'New Category', got %v", cat["name"])
	}
}

func TestCategoriesGet(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(t, h, "GET", "/api/v1/categories/1", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var env apiEnvelope
	json.Unmarshal(rec.Body.Bytes(), &env)

	var cat map[string]any
	json.Unmarshal(env.Data, &cat)

	if cat["name"] != "Seed Category" {
		t.Fatalf("expected Seed Category, got %v", cat["name"])
	}
}

func TestCategoriesUpdate(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	body := `{"name":"Updated Category"}`
	rec := doReq(t, h, "PATCH", "/api/v1/categories/1", []byte(body))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// fetch updated
	rec2 := doReq(t, h, "GET", "/api/v1/categories/1", nil)

	var env apiEnvelope
	json.Unmarshal(rec2.Body.Bytes(), &env)

	var cat map[string]any
	json.Unmarshal(env.Data, &cat)

	if cat["name"] != "Updated Category" {
		t.Fatalf("expected updated name, got %v", cat["name"])
	}
}

func TestCategoriesDelete(t *testing.T) {
	h, db := newCategoryAPI(t)
	defer db.Close()

	rec := doReq(t, h, "DELETE", "/api/v1/categories/1", nil)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}

	// category must not exist anymore
	rec2 := doReq(t, h, "GET", "/api/v1/categories/1", nil)

	if rec2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deletion, got %d", rec2.Code)
	}
}
