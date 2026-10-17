package response

import (
	"fmt"
	"math"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Meta carries pagination metadata alongside list data.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalData  int64 `json:"total_data"`
	TotalPages int   `json:"total_pages"`
}

// NewMeta computes total pages from total rows and limit.
func NewMeta(page, perPage int, totalData int64) *Meta {
	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(totalData) / float64(perPage)))
	}
	return &Meta{
		Page:       page,
		PerPage:    perPage,
		TotalData:  totalData,
		TotalPages: totalPages,
	}
}

// PaginationParams holds clamped pagination and sorting options.
type PaginationParams struct {
	Page      int    `json:"page" query:"page"`
	PerPage   int    `json:"per_page" query:"per_page"`
	SortBy    string `json:"sort_by" query:"sort_by"`
	SortOrder string `json:"sort_order" query:"sort_order"`
}

// Offset returns the SQL OFFSET value.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Paginate returns a GORM scope function that applies OFFSET and LIMIT.
func (p PaginationParams) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.Offset()).Limit(p.PerPage)
	}
}

// Scope is an alias for Paginate, allowing db.Scopes(params.Scope()).
func (p PaginationParams) Scope() func(db *gorm.DB) *gorm.DB {
	return p.Paginate()
}

// Sort returns a GORM scope for ordering. If SortBy is empty, defaultSort is used.
func (p PaginationParams) Sort(defaultSort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if p.SortBy == "" {
			if defaultSort != "" {
				return db.Order(defaultSort)
			}
			return db
		}

		order := strings.ToLower(p.SortOrder)
		if order != "asc" && order != "desc" {
			order = "asc"
		}

		return db.Order(fmt.Sprintf("%s %s", p.SortBy, order))
	}
}

// Paginate returns a GORM scope function for the provided PaginationParams.
func Paginate(p PaginationParams) func(db *gorm.DB) *gorm.DB {
	return p.Paginate()
}

// PaginateScope returns a GORM scope given raw page and perPage integers.
func PaginateScope(page, perPage int) func(db *gorm.DB) *gorm.DB {
	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	offset := (page - 1) * perPage
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(perPage)
	}
}

// ParsePagination reads page, per_page (or limit), sort_by, and sort_order with safe defaults and limits.
func ParsePagination(c fiber.Ctx) PaginationParams {
	page := fiber.Query(c, "page", DefaultPage)

	perPage := fiber.Query(c, "per_page", 0)
	if perPage == 0 {
		perPage = fiber.Query(c, "limit", DefaultPerPage)
	}

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	sortBy := fiber.Query(c, "sort_by", "")
	if sortBy == "" {
		sortBy = fiber.Query(c, "sort", "")
	}

	sortOrder := fiber.Query(c, "sort_order", "")
	if sortOrder == "" {
		sortOrder = fiber.Query(c, "order", "")
	}

	return PaginationParams{
		Page:      page,
		PerPage:   perPage,
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}
}
