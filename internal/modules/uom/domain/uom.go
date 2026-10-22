package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// UoMCategory defines the measurement category for a unit.
type UoMCategory string

const (
	UoMCategoryWeight UoMCategory = "WEIGHT" // Gram, Kilogram, Milligram, Ounce
	UoMCategoryVolume UoMCategory = "VOLUME" // Milliliter, Liter, Gallon
	UoMCategoryUnit   UoMCategory = "UNIT"   // Piece, Pack, Box, Can, Bottle, Dozen
)

// IsValid checks whether the category is a recognized UoMCategory.
func (c UoMCategory) IsValid() bool {
	switch c {
	case UoMCategoryWeight, UoMCategoryVolume, UoMCategoryUnit:
		return true
	default:
		return false
	}
}

// UoM represents a unit of measurement entity.
type UoM struct {
	ID             string      `json:"id"`
	OrganizationID *string     `json:"organization_id,omitempty"` // nil for system global units
	Name           string      `json:"name"`
	Symbol         string      `json:"symbol"`
	Category       UoMCategory `json:"category"`
	DecimalPlaces  int         `json:"decimal_places"`
	IsSystem       bool        `json:"is_system"`
	Status         string      `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	DeletedAt      *time.Time  `json:"deleted_at,omitempty"`
}

// NewCustomUoM creates a new tenant-specific custom UoM entity.
func NewCustomUoM(orgID, name, symbol string, category UoMCategory, decimals int) *UoM {
	now := time.Now().UTC()
	return &UoM{
		ID:             uuid.NewString(),
		OrganizationID: &orgID,
		Name:           name,
		Symbol:         symbol,
		Category:       category,
		DecimalPlaces:  decimals,
		IsSystem:       false,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// NewSystemUoM creates a built-in immutable system UoM entity.
func NewSystemUoM(id, name, symbol string, category UoMCategory, decimals int) *UoM {
	now := time.Now().UTC()
	return &UoM{
		ID:             id,
		OrganizationID: nil,
		Name:           name,
		Symbol:         symbol,
		Category:       category,
		DecimalPlaces:  decimals,
		IsSystem:       true,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UoMSummary provides a lightweight DTO for foreign domain cross-referencing.
type UoMSummary struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Symbol        string      `json:"symbol"`
	Category      UoMCategory `json:"category"`
	DecimalPlaces int         `json:"decimal_places"`
}

// ToSummary converts the UoM entity into a foreign-friendly summary.
func (u *UoM) ToSummary() UoMSummary {
	return UoMSummary{
		ID:            u.ID,
		Name:          u.Name,
		Symbol:        u.Symbol,
		Category:      u.Category,
		DecimalPlaces: u.DecimalPlaces,
	}
}

// UoMFilter specifies filtering criteria for querying UoMs.
type UoMFilter struct {
	Search   string
	Category *UoMCategory
	Status   *string
}

// UoMRepository defines the persistence contract for UoM entities.
type UoMRepository interface {
	GetByID(ctx context.Context, id string) (*UoM, error)
	GetByIDs(ctx context.Context, ids []string) ([]*UoM, error)
	GetBySymbol(ctx context.Context, orgID *string, symbol string) (*UoM, error)
	List(ctx context.Context, orgID string, filter UoMFilter, pagination response.PaginationParams) ([]*UoM, int64, error)
	ListLookup(ctx context.Context, orgID string, category *UoMCategory) ([]*UoM, error)
	Create(ctx context.Context, entity *UoM) error
	Update(ctx context.Context, entity *UoM) error
	Delete(ctx context.Context, id string) error
}
