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
	BrandID        *string    `json:"brand_id,omitempty"`
	OutletID       *string    `json:"outlet_id,omitempty"`
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

// UserRepository defines persistence contracts for User entities.
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByIDs(ctx context.Context, ids []string) ([]*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, orgID, username string) (*User, error)
	GetByOutletAndUsername(ctx context.Context, outletID, username string) (*User, error)
	ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*User, int64, error)
	ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*User, int64, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}
