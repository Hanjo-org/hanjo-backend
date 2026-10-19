package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
)

// RoleModel maps to the roles table in PostgreSQL.
type RoleModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID *string        `gorm:"column:organization_id;type:uuid;index"`
	Code           string         `gorm:"column:code;type:varchar(50);not null"`
	Name           string         `gorm:"column:name;type:varchar(100);not null"`
	Description    string         `gorm:"column:description;type:text"`
	Tier           string         `gorm:"column:tier;type:varchar(20);not null;index"`
	IsSystem       bool           `gorm:"column:is_system;not null;default:false"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE';index"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (RoleModel) TableName() string {
	return "roles"
}

func (m *RoleModel) ToDomain() *domain.Role {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.Role{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		Code:           m.Code,
		Name:           m.Name,
		Description:    m.Description,
		Tier:           domain.RoleTier(m.Tier),
		IsSystem:       m.IsSystem,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromRoleDomain(d *domain.Role) *RoleModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &RoleModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		Code:           d.Code,
		Name:           d.Name,
		Description:    d.Description,
		Tier:           string(d.Tier),
		IsSystem:       d.IsSystem,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

// RolePermissionModel maps to the role_permissions junction table.
type RolePermissionModel struct {
	RoleID       string    `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionID string    `gorm:"column:permission_id;type:uuid;primaryKey"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (RolePermissionModel) TableName() string {
	return "role_permissions"
}
