package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
)

type BrandModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID string         `gorm:"column:organization_id;type:uuid;not null;index"`
	Name           string         `gorm:"column:name;type:varchar(255);not null"`
	Slug           string         `gorm:"column:slug;type:varchar(100);not null"`
	LogoURL        *string        `gorm:"column:logo_url;type:text"`
	Tagline        *string        `gorm:"column:tagline;type:varchar(255)"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (BrandModel) TableName() string {
	return "brands"
}

func (m *BrandModel) ToDomain() *domain.Brand {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.Brand{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		Name:           m.Name,
		Slug:           m.Slug,
		LogoURL:        m.LogoURL,
		Tagline:        m.Tagline,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromBrandDomain(d *domain.Brand) *BrandModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &BrandModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		Name:           d.Name,
		Slug:           d.Slug,
		LogoURL:        d.LogoURL,
		Tagline:        d.Tagline,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}
