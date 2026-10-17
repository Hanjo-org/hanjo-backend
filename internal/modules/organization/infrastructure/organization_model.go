package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
)

type OrganizationModel struct {
	ID        string         `gorm:"column:id;type:uuid;primaryKey"`
	Name      string         `gorm:"column:name;type:varchar(255);not null"`
	Slug      string         `gorm:"column:slug;type:varchar(100);not null"`
	Status    string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (OrganizationModel) TableName() string {
	return "organizations"
}

func (m *OrganizationModel) ToDomain() *domain.Organization {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.Organization{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func FromOrganizationDomain(d *domain.Organization) *OrganizationModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &OrganizationModel{
		ID:        d.ID,
		Name:      d.Name,
		Slug:      d.Slug,
		Status:    d.Status,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
		DeletedAt: deletedAt,
	}
}
