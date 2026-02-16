package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAPIDraftCreate_ManualImageOnlyMultipartAndRestore(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	rec := multipartRequest(
		t,
		h,
		http.MethodPost,
		"/api/v1/posts/draft",
		token,
		map[string]string{
			"title":  "Draft with image only",
			"body":   "",
			"manual": "true",
		},
		[]int64{1},
		"draft.jpg",
		sampleJPEGBytes,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	draftCreate := decodeEnvelopeDataMap(t, rec)
	draftID := int64(draftCreate["id"].(float64))

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts/draft", nil)
	getReq.Header.Set("Cookie", "session_token="+token)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft get, got %d body=%s", getRec.Code, getRec.Body.String())
	}

	draft := decodeEnvelopeDataMap(t, getRec)
	if int64(draft["id"].(float64)) != draftID {
		t.Fatalf("expected draft id %d, got %v", draftID, draft["id"])
	}
	if draft["title"] != "Draft with image only" {
		t.Fatalf("expected restored title, got %v", draft["title"])
	}
	if draft["body"] != "" {
		t.Fatalf("expected empty body for image-only draft, got %v", draft["body"])
	}
	rawURL, ok := draft["image_url"]
	if !ok || rawURL == nil {
		t.Fatalf("expected image_url on restored draft, got %v", draft)
	}
	imageURL := rawURL.(string)
	if strings.TrimSpace(imageURL) == "" {
		t.Fatalf("expected non-empty image_url on restored draft")
	}
	t.Cleanup(func() { cleanupUploadedFromImageURL(t, imageURL) })
}

func TestAPIDraftUpdate_ManualImageOnlyMultipart(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	// create a starting draft via JSON
	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/posts/draft",
		strings.NewReader(`{"title":"Initial Draft","body":"Initial body","category_ids":[1],"manual":true}`),
	)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Cookie", "session_token="+token)
	createRec := httptest.NewRecorder()
	h.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on initial draft create, got %d body=%s", createRec.Code, createRec.Body.String())
	}
	createData := decodeEnvelopeDataMap(t, createRec)
	draftID := int64(createData["id"].(float64))

	updateRec := multipartRequest(
		t,
		h,
		http.MethodPut,
		fmt.Sprintf("/api/v1/posts/draft/%d", draftID),
		token,
		map[string]string{
			"title":  "Updated Draft",
			"body":   "",
			"manual": "true",
		},
		[]int64{1},
		"updated.png",
		samplePNGBytes,
	)
	if updateRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts/draft", nil)
	getReq.Header.Set("Cookie", "session_token="+token)
	getRec := httptest.NewRecorder()
	h.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 on draft get, got %d body=%s", getRec.Code, getRec.Body.String())
	}
	draft := decodeEnvelopeDataMap(t, getRec)

	if int64(draft["id"].(float64)) != draftID {
		t.Fatalf("expected updated draft id %d, got %v", draftID, draft["id"])
	}
	if draft["title"] != "Updated Draft" {
		t.Fatalf("expected updated title, got %v", draft["title"])
	}
	if draft["body"] != "" {
		t.Fatalf("expected empty body after image-only update, got %v", draft["body"])
	}
	rawURL, ok := draft["image_url"]
	if !ok || rawURL == nil {
		t.Fatalf("expected image_url after draft image update, got %v", draft)
	}
	imageURL := rawURL.(string)
	t.Cleanup(func() { cleanupUploadedFromImageURL(t, imageURL) })
}

func TestAPIDraftCreate_RejectsUnsupportedImageType(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	rec := multipartRequest(
		t,
		h,
		http.MethodPost,
		"/api/v1/posts/draft",
		token,
		map[string]string{
			"title":  "Bad draft image",
			"body":   "Has text",
			"manual": "true",
		},
		[]int64{1},
		"bad.gif",
		sampleTextBytes,
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}

	apiErr := decodeErrorEnvelope(t, rec)
	if apiErr == nil {
		t.Fatalf("expected error envelope, got body=%s", rec.Body.String())
	}
	if apiErr.Message != "unsupported image type" {
		t.Fatalf("expected unsupported image type, got %q", apiErr.Message)
	}
}

func TestAPIDraftCreate_RejectsExtensionMimeMismatch(t *testing.T) {
	h, db := newTestAPI(t)
	defer db.Close()

	token := loginAndGetToken(t, h, "testuser", "password123")

	rec := multipartRequest(
		t,
		h,
		http.MethodPost,
		"/api/v1/posts/draft",
		token,
		map[string]string{
			"title":  "Mismatch draft image",
			"body":   "Body",
			"manual": "true",
		},
		[]int64{1},
		"draft.jpg",
		sampleGIFBytes,
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
	apiErr := decodeErrorEnvelope(t, rec)
	if apiErr == nil {
		t.Fatalf("expected error envelope, got body=%s", rec.Body.String())
	}
	if apiErr.Message != "unsupported image type" {
		t.Fatalf("expected unsupported image type, got %q", apiErr.Message)
	}
}

func TestCreatePostTemplate_HasImagePreviewControls(t *testing.T) {
	path := "../../web/templates/create-post.html"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read template %s: %v", path, err)
	}
	s := string(b)

	want := []string{
		`id="image"`,
		`name="image"`,
		`accept="image/jpeg,image/png,image/gif"`,
		`id="image-preview"`,
		`id="image-name"`,
		`id="image-clear"`,
	}
	for _, token := range want {
		if !strings.Contains(s, token) {
			t.Fatalf("expected template to include %q", token)
		}
	}
}

func TestCreatePostJS_ConditionallyRendersPostImageMarkup(t *testing.T) {
	path := "../../web/static/js/posts.js"
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read js %s: %v", path, err)
	}
	s := string(b)

	if !strings.Contains(s, "const imageMarkup = imageUrl") {
		t.Fatalf("expected posts.js to define conditional image markup")
	}
	if !strings.Contains(s, `: ""`) {
		t.Fatalf("expected posts.js to omit image container when no image URL")
	}
}
