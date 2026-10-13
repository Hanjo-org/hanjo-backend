package domain

import (
	"errors"
	"fmt"
	"net/http"
)

// DomainError represents a business rule violation or domain-specific error.
type DomainError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Common domain errors
var (
	ErrNotFound = &DomainError{
		Code:       "NOT_FOUND",
		Message:    "resource not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAlreadyExists = &DomainError{
		Code:       "ALREADY_EXISTS",
		Message:    "resource already exists",
		HTTPStatus: http.StatusConflict,
	}

	ErrUnauthorized = &DomainError{
		Code:       "UNAUTHORIZED",
		Message:    "authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &DomainError{
		Code:       "FORBIDDEN",
		Message:    "insufficient permissions",
		HTTPStatus: http.StatusForbidden,
	}

	ErrInvalidInput = &DomainError{
		Code:       "INVALID_INPUT",
		Message:    "invalid input provided",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrConflict = &DomainError{
		Code:       "CONFLICT",
		Message:    "operation conflicts with current state",
		HTTPStatus: http.StatusConflict,
	}

	ErrInternalError = &DomainError{
		Code:       "INTERNAL_ERROR",
		Message:    "an internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrOptimisticLock = &DomainError{
		Code:       "OPTIMISTIC_LOCK",
		Message:    "resource was modified by another process",
		HTTPStatus: http.StatusConflict,
	}
)

// NewDomainError creates a custom domain error with the given code, message and HTTP status.
func NewDomainError(code, message string, httpStatus int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// NewValidationError creates a domain error for a specific validation failure.
func NewValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:       "VALIDATION_ERROR",
		Message:    fmt.Sprintf("%s: %s", field, message),
		HTTPStatus: http.StatusBadRequest,
	}
}

// NewNotFoundError creates a not-found error for a specific resource type.
func NewNotFoundError(resourceType string) *DomainError {
	return &DomainError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s not found", resourceType),
		HTTPStatus: http.StatusNotFound,
	}
}

// IsDomainError checks if the given error is a DomainError.
func IsDomainError(err error) bool {
	var domainErr *DomainError
	return errors.As(err, &domainErr)
}

// GetDomainError extracts the DomainError from an error chain.
func GetDomainError(err error) (*DomainError, bool) {
	var domainErr *DomainError
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}
