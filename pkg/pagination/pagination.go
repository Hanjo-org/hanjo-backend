package pagination

import (
	"math"
	"net/http"
	"strconv"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params holds pagination parameters.
type Params struct {
	Page    int
	PerPage int
}

// Offset returns the database offset for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the number of items per page.
func (p Params) Limit() int {
	return p.PerPage
}

// Result holds pagination result metadata.
type Result struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// NewResult creates a pagination result from params and total count.
func NewResult(params Params, total int64) Result {
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	return Result{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// FromRequest extracts pagination parameters from an HTTP request's query params.
func FromRequest(r *http.Request) Params {
	page := DefaultPage
	perPage := DefaultPerPage

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	if pp := r.URL.Query().Get("per_page"); pp != "" {
		if val, err := strconv.Atoi(pp); err == nil && val > 0 {
			perPage = val
			if perPage > MaxPerPage {
				perPage = MaxPerPage
			}
		}
	}

	return Params{Page: page, PerPage: perPage}
}
