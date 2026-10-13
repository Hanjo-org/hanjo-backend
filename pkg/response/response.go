package response

import (
	"encoding/json"
	"net/http"

	"github.com/hanjo-erp/hanjo/internal/common/domain"
)

// APIResponse is the standard JSON response envelope.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorBody holds error details in the response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta holds pagination and other metadata.
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// JSON writes a success JSON response.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
	})
}

// JSONWithMeta writes a success JSON response with pagination metadata.
func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta *Meta) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Created writes a 201 Created response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error writes an error JSON response.
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIResponse{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// DomainError writes an error response from a domain error.
func DomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := domain.GetDomainError(err); ok {
		Error(w, domainErr.HTTPStatus, domainErr.Code, domainErr.Message)
		return
	}
	// Fallback for non-domain errors
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found response.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred")
}
