package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
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

func WriteError(w http.ResponseWriter, r *http.Request, err *APIError) {
	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		tmplPath := fmt.Sprintf("./web/errors/%d.html", err.Status)
		tmpl, tmplErr := template.ParseFiles(tmplPath)
		if tmplErr != nil {
			http.Error(w, err.Message, err.Status)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(err.Status)
		tmpl.Execute(w, err)
	} else {
		//Fallback to json
		writeJSON(w, err.Status, &APIResponse{
			Error: err,
		})
	}
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
