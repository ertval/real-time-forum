package handlers

import (
	"encoding/json"
	"net/http"
)

// ------------------------------------------------------------
// API RESPONSE ENVELOPE
// ------------------------------------------------------------

// APIResponse is the unified JSON envelope for all API responses.
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Meta  any       `json:"meta,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

// APIError represents a standardized API error.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"` // HTTP status code (not serialized)
}

// ------------------------------------------------------------
// SUCCESS RESPONSES
// ------------------------------------------------------------

func WriteOK(w http.ResponseWriter, data any, meta any) {
	resp := &APIResponse{
		Data: data,
	}
	if meta != nil {
		resp.Meta = meta
	}
	writeJSON(w, http.StatusOK, resp)
}

func WriteCreated(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusCreated, &APIResponse{
		Data: data,
	})
}

func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ------------------------------------------------------------
// ERROR RESPONSES
// ------------------------------------------------------------

func WriteError(w http.ResponseWriter, err *APIError) {
	writeJSON(w, err.Status, &APIResponse{
		Error: err,
	})
}

func NewError(code, message string, status int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// ------------------------------------------------------------
// INTERNAL JSON WRITER
// ------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, body *APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
