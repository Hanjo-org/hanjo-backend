package gormscope

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// FilterLike returns a GORM scope that performs a case-insensitive LIKE (or ILIKE) search.
// If val is empty, the scope is a no-op.
func FilterLike(column string, val string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" {
			return db
		}
		// Uses ILIKE for PostgreSQL (fallback compatible with LIKE)
		return db.Where(fmt.Sprintf("%s ILIKE ?", column), "%"+trimmed+"%")
	}
}

// FilterExact returns a GORM scope matching exact equality if val is non-nil.
func FilterExact[T comparable](column string, val *T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if val == nil {
			return db
		}
		return db.Where(fmt.Sprintf("%s = ?", column), *val)
	}
}

// FilterIn returns a GORM scope for WHERE IN if the slice contains elements.
func FilterIn[T any](column string, slice []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(slice) == 0 {
			return db
		}
		return db.Where(fmt.Sprintf("%s IN (?)", column), slice)
	}
}

// FilterNotIn returns a GORM scope for WHERE NOT IN if the slice contains elements.
func FilterNotIn[T any](column string, slice []T) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if len(slice) == 0 {
			return db
		}
		return db.Where(fmt.Sprintf("%s NOT IN (?)", column), slice)
	}
}

// FilterDateRange returns a GORM scope filtering between from and to timestamps.
func FilterDateRange(column string, from, to *time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if from != nil && !from.IsZero() {
			db = db.Where(fmt.Sprintf("%s >= ?", column), *from)
		}
		if to != nil && !to.IsZero() {
			db = db.Where(fmt.Sprintf("%s <= ?", column), *to)
		}
		return db
	}
}

// FilterActive filters records where is_active is true.
// Default column name is "is_active".
func FilterActive(column ...string) func(db *gorm.DB) *gorm.DB {
	col := "is_active"
	if len(column) > 0 && column[0] != "" {
		col = column[0]
	}
	return func(db *gorm.DB) *gorm.DB {
		return db.Where(fmt.Sprintf("%s = ?", col), true)
	}
}
