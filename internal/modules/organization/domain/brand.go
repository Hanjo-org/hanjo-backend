package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Brand represents the second tier (Tier 2: Coffee Shop Brand / Business Concept).
type Brand struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	LogoURL        *string    `json:"logo_url,omitempty"`
	Tagline        *string    `json:"tagline,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewBrand creates a new Brand entity initialized with default UUID and timestamps.
func NewBrand(orgID, name, slug string) *Brand {
	now := time.Now().UTC()
	return &Brand{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// BrandRepository defines persistence contracts for Brand entities.
type BrandRepository interface {
	GetByID(ctx context.Context, id string) (*Brand, error)
	GetBySlug(ctx context.Context, orgID, slug string) (*Brand, error)
	ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*Brand, int64, error)
	Create(ctx context.Context, brand *Brand) error
	Update(ctx context.Context, brand *Brand) error
	Delete(ctx context.Context, id string) error
}
