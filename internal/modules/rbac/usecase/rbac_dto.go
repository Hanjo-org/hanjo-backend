package usecase

import (
	"time"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
)

// CreateRoleInput defines parameters for creating a new custom role.
type CreateRoleInput struct {
	OrganizationID string
	Code           string
	Name           string
	Description    string
	Tier           string
	PermissionIDs  []string
}

// UpdateRoleInput defines parameters for updating an existing role.
type UpdateRoleInput struct {
	Name          *string
	Description   *string
	PermissionIDs []string
	Status        *string
}

// RoleResponse represents the full role payload including permissions.
type RoleResponse struct {
	ID             string                `json:"id"`
	OrganizationID *string               `json:"organization_id,omitempty"`
	Code           string                `json:"code"`
	Name           string                `json:"name"`
	Description    string                `json:"description"`
	Tier           string                `json:"tier"`
	IsSystem       bool                  `json:"is_system"`
	Status         string                `json:"status"`
	Permissions    []*PermissionResponse `json:"permissions"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
}

// RoleLookupResponse represents a safe, minimal role representation for UI comboboxes.
type RoleLookupResponse struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

// PermissionResponse represents an individual permission item.
type PermissionResponse struct {
	ID          string `json:"id"`
	Code        string `json:"code"`
	Group       string `json:"group"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PermissionGroupResponse groups permissions by module for UI checkbox matrix rendering.
type PermissionGroupResponse struct {
	Group       string                `json:"group"`
	GroupName   string                `json:"group_name"`
	Permissions []*PermissionResponse `json:"permissions"`
}

// Helper converters
func ToPermissionResponse(p *domain.Permission) *PermissionResponse {
	if p == nil {
		return nil
	}
	return &PermissionResponse{
		ID:          p.ID,
		Code:        p.Code,
		Group:       p.Group,
		Name:        p.Name,
		Description: p.Description,
	}
}

func ToRoleLookupResponse(l *domain.RoleLookup) *RoleLookupResponse {
	if l == nil {
		return nil
	}
	return &RoleLookupResponse{
		ID:   l.ID,
		Code: l.Code,
		Name: l.Name,
		Tier: string(l.Tier),
	}
}
