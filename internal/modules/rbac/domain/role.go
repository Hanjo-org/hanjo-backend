package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// RoleTier specifies the hierarchical boundary of a role.
type RoleTier string

const (
	RoleTierOrganization RoleTier = "ORGANIZATION"
	RoleTierBrand        RoleTier = "BRAND"
	RoleTierOutlet       RoleTier = "OUTLET"
)

// Role represents a system default or organization custom role.
type Role struct {
	ID             string     `json:"id"`
	OrganizationID *string    `json:"organization_id"` // nil for system global roles
	Code           string     `json:"code"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	Tier           RoleTier   `json:"tier"`
	IsSystem       bool       `json:"is_system"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewCustomRole initializes a new custom role entity for an organization.
func NewCustomRole(orgID, code, name, description string, tier RoleTier) *Role {
	now := time.Now().UTC()
	return &Role{
		ID:             uuid.NewString(),
		OrganizationID: &orgID,
		Code:           code,
		Name:           name,
		Description:    description,
		Tier:           tier,
		IsSystem:       false,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// RoleSummary provides a lightweight summary of a role for embedding in foreign domain responses (e.g. User).
type RoleSummary struct {
	ID   string   `json:"id"`
	Code string   `json:"code"`
	Name string   `json:"name"`
	Tier RoleTier `json:"tier"`
}

// RoleWithPermissions represents a role along with its granted permissions.
type RoleWithPermissions struct {
	*Role
	Permissions []*Permission `json:"permissions"`
}

// RoleLookup represents a safe, minimal role representation for UI comboboxes and dropdowns.
type RoleLookup struct {
	ID   string   `json:"id"`
	Code string   `json:"code"`
	Name string   `json:"name"`
	Tier RoleTier `json:"tier"`
}

// RoleFilter defines query criteria for listing roles.
type RoleFilter struct {
	OrganizationID *string
	Tier           *RoleTier
	Status         *string
	Search         *string
}

// RoleRepository defines persistence contracts for roles and role-permission associations.
type RoleRepository interface {
	GetByID(ctx context.Context, id string) (*Role, error)
	GetByIDs(ctx context.Context, ids []string) ([]*Role, error)
	GetByCode(ctx context.Context, orgID *string, code string) (*Role, error)
	List(ctx context.Context, filter RoleFilter, params response.PaginationParams) ([]*Role, int64, error)
	GetLookup(ctx context.Context, orgID string, tier *RoleTier, params response.PaginationParams) ([]*RoleLookup, int64, error)
	Create(ctx context.Context, role *Role, permissionIDs []string) error
	Update(ctx context.Context, role *Role, permissionIDs []string) error
	Delete(ctx context.Context, id string) error
	GetAllRolePermissionsMapping(ctx context.Context) (map[string][]string, error)
	GetAllRoleCodes(ctx context.Context) (map[string]string, error)
}
