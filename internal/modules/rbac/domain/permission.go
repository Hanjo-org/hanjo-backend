package domain

import (
	"context"
	"time"
)

// Permission represents an atomic technical capability within the system (e.g. 'pos:order:void').
type Permission struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Group       string    `json:"group"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// PermissionFilter specifies criteria for querying permissions.
type PermissionFilter struct {
	Group  *string
	Search *string
}

// PermissionRepository defines persistence contracts for system permissions.
type PermissionRepository interface {
	GetByID(ctx context.Context, id string) (*Permission, error)
	GetByCode(ctx context.Context, code string) (*Permission, error)
	GetByCodes(ctx context.Context, codes []string) ([]*Permission, error)
	List(ctx context.Context, filter PermissionFilter) ([]*Permission, error)
	ListByRoleID(ctx context.Context, roleID string) ([]*Permission, error)
	ListByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]*Permission, error)
}
