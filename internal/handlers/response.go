// internal/handlers/response.go
package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strings"
)

/* ------------------------------------------------------------
   API RESPONSE ENVELOPE
   ------------------------------------------------------------ */

// APIResponse is the unified JSON envelope for all API responses.
type APIResponse struct {
	Data  any       `json:"data,omitempty"`
	Meta  any       `json:"meta,omitempty"`
	Error *APIError `json:"error,omitempty"`
}

type ErrorPageData struct {
	Code    int
	Title   string
	Message string
}

/* ------------------------------------------------------------
   SUCCESS RESPONSES
   ------------------------------------------------------------ */

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

/* ------------------------------------------------------------
   ERROR RESPONSES
   ------------------------------------------------------------ */

func WriteError(w http.ResponseWriter, r *http.Request, err *APIError) {
	acceptsHTML := strings.Contains(r.Header.Get("Accept"), "text/html")
	//Custom error page shouldn't be served for forbidden and unauthorized http errors
	isAuthError := err.Status == http.StatusUnauthorized || err.Status == http.StatusForbidden
	if acceptsHTML && !isAuthError {
		tmpl, tmplErr := template.ParseFiles("./web/errors/error.html")
		if tmplErr != nil {
			http.Error(w, err.Message, err.Status)
			return
		}
		data := ErrorPageData{
			Code:    err.Status,
			Title:   statusTitle(err.Status),
			Message: err.Message,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(err.Status)
		if execErr := tmpl.Execute(w, data); execErr != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}
	//Fallback to JSON
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

/* ------------------------------------------------------------
   INTERNAL JSON WRITER
   ------------------------------------------------------------ */

func writeJSON(w http.ResponseWriter, status int, body *APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

/* ------------------------------------------------------------
	Writeerror Helpers
   ------------------------------------------------------------ */

func statusTitle(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusNotFound:
		return "Not Found"
	case http.StatusInternalServerError:
		return "Internal Server Error"
	default:
		return "Error"
	}
}
