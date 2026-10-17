package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Organization represents the highest tier (Tier 1: Holding / Management Group).
type Organization struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// NewOrganization creates a new Organization entity initialized with default UUID and timestamps.
func NewOrganization(name, slug string) *Organization {
	now := time.Now().UTC()
	return &Organization{
		ID:        uuid.NewString(),
		Name:      name,
		Slug:      slug,
		Status:    "ACTIVE",
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// OrganizationRepository defines persistence contracts for Organization entities.
type OrganizationRepository interface {
	GetByID(ctx context.Context, id string) (*Organization, error)
	GetBySlug(ctx context.Context, slug string) (*Organization, error)
	List(ctx context.Context, params response.PaginationParams) ([]*Organization, int64, error)
	Create(ctx context.Context, org *Organization) error
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id string) error
}
