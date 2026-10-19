package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// User represents a system identity (staff, cashier, barista, manager, or owner).
type User struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	RoleID         *string    `json:"role_id,omitempty"`
	BrandID        *string    `json:"brand_id"`
	OutletID       *string    `json:"outlet_id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Password       string     `json:"-"`
	Pin            *string    `json:"-"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewUser creates a new User entity initialized with a UUID and timestamps.
func NewUser(orgID, username, email, password string) *User {
	now := time.Now().UTC()
	return &User{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Username:       username,
		Email:          email,
		Password:       password,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// OrganizationSummary represents minimal organization details embedded in user queries.
type OrganizationSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// BrandSummary represents minimal brand details embedded in user queries.
type BrandSummary struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Slug    string  `json:"slug"`
	LogoURL *string `json:"logo_url"`
}

// OutletSummary represents minimal outlet details embedded in user queries.
type OutletSummary struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Code    string  `json:"code"`
	Address string  `json:"address"`
}

// RoleSummary represents minimal role details embedded in user queries.
type RoleSummary struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

// UserWithRelations embeds related organization, brand, outlet, and role summaries with a User.
type UserWithRelations struct {
	*User
	Organization *OrganizationSummary `json:"organization"`
	Brand        *BrandSummary        `json:"brand"`
	Outlet       *OutletSummary       `json:"outlet"`
	Role         *RoleSummary         `json:"role"`
}

// UserLookup represents a lightweight reference of an active user for UI comboboxes and dropdowns.
type UserLookup struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	OutletID *string `json:"outlet_id"`
	BrandID  *string `json:"brand_id"`
}

// UserFilter defines query criteria for listing users.
type UserFilter struct {
	OrganizationID string
	BrandID        *string
	OutletID       *string
	Status         *string
	Search         *string
}

// UserRepository defines persistence contracts for User entities.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByIDs(ctx context.Context, ids []string) ([]*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, orgID, username string) (*User, error)
	GetByOutletAndUsername(ctx context.Context, outletID, username string) (*User, error)
	List(ctx context.Context, filter UserFilter, params response.PaginationParams) ([]*User, int64, error)
	ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*User, int64, error)
	ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*User, int64, error)
	GetLookup(ctx context.Context, filter UserFilter, params response.PaginationParams) ([]*UserLookup, int64, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
