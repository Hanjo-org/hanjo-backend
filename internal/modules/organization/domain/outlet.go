package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Outlet represents the third tier (Tier 3: Physical Store / Branch / POS Location).
type Outlet struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	BrandID        string     `json:"brand_id"`
	Name           string     `json:"name"`
	Code           string     `json:"code"`
	Address        string     `json:"address"`
	Phone          *string    `json:"phone,omitempty"`
	Latitude       *float64   `json:"latitude,omitempty"`
	Longitude      *float64   `json:"longitude,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewOutlet creates a new Outlet entity initialized with default UUID and timestamps.
func NewOutlet(orgID, brandID, name, code, address string) *Outlet {
	now := time.Now().UTC()
	return &Outlet{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		BrandID:        brandID,
		Name:           name,
		Code:           code,
		Address:        address,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// OutletRepository defines persistence contracts for Outlet entities.
type OutletRepository interface {
	GetByID(ctx context.Context, id string) (*Outlet, error)
	GetByIDs(ctx context.Context, ids []string) ([]*Outlet, error)
	GetByCode(ctx context.Context, orgID, code string) (*Outlet, error)
	ListByBrand(ctx context.Context, brandID string, params response.PaginationParams) ([]*Outlet, int64, error)
	ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*Outlet, int64, error)
	Create(ctx context.Context, outlet *Outlet) error
	Update(ctx context.Context, outlet *Outlet) error
	Delete(ctx context.Context, id string) error
}
