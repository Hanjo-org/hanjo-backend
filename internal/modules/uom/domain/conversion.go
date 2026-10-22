package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UoMConversion represents a conversion ratio between two units of measurement.
type UoMConversion struct {
	ID             string     `json:"id"`
	OrganizationID *string    `json:"organization_id,omitempty"` // nil for global system standards
	FromUoMID      string     `json:"from_uom_id"`
	ToUoMID        string     `json:"to_uom_id"`
	Factor         float64    `json:"factor"` // multiplier: 1 from_unit = factor * to_unit
	IsSystem       bool       `json:"is_system"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// NewConversion initializes a new unit conversion entity.
func NewConversion(orgID *string, fromUoMID, toUoMID string, factor float64, isSystem bool) *UoMConversion {
	now := time.Now().UTC()
	return &UoMConversion{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		FromUoMID:      fromUoMID,
		ToUoMID:        toUoMID,
		Factor:         factor,
		IsSystem:       isSystem,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// UoMConversionRepository defines the persistence contract for conversions.
type UoMConversionRepository interface {
	GetByID(ctx context.Context, id string) (*UoMConversion, error)
	GetByPair(ctx context.Context, fromUoMID, toUoMID string) (*UoMConversion, error)
	ListByUoMID(ctx context.Context, uomID string) ([]*UoMConversion, error)
	List(ctx context.Context, orgID string) ([]*UoMConversion, error)
	Create(ctx context.Context, conversion *UoMConversion) error
	Delete(ctx context.Context, id string) error
}
