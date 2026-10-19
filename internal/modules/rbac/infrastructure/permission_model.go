package infrastructure

import (
	"time"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
)

// PermissionModel maps to the permissions table in PostgreSQL.
type PermissionModel struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey"`
	Code        string    `gorm:"column:code;type:varchar(100);not null;uniqueIndex"`
	Group       string    `gorm:"column:group;type:varchar(50);not null;index"`
	Name        string    `gorm:"column:name;type:varchar(150);not null"`
	Description string    `gorm:"column:description;type:text"`
	CreatedAt   time.Time `gorm:"column:created_at;not null"`
}

func (PermissionModel) TableName() string {
	return "permissions"
}

func (m *PermissionModel) ToDomain() *domain.Permission {
	if m == nil {
		return nil
	}
	return &domain.Permission{
		ID:          m.ID,
		Code:        m.Code,
		Group:       m.Group,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func FromPermissionDomain(d *domain.Permission) *PermissionModel {
	if d == nil {
		return nil
	}
	return &PermissionModel{
		ID:          d.ID,
		Code:        d.Code,
		Group:       d.Group,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
	}
}
