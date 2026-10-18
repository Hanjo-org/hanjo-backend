package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
)

type OutletModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID string         `gorm:"column:organization_id;type:uuid;not null;index"`
	BrandID        string         `gorm:"column:brand_id;type:uuid;not null;index"`
	Name           string         `gorm:"column:name;type:varchar(255);not null"`
	Code           string         `gorm:"column:code;type:varchar(50);not null"`
	Address        string         `gorm:"column:address;type:text;not null"`
	Phone          *string        `gorm:"column:phone;type:varchar(50)"`
	Latitude       *float64       `gorm:"column:latitude;type:double precision"`
	Longitude      *float64       `gorm:"column:longitude;type:double precision"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (OutletModel) TableName() string {
	return "outlets"
}

func (m *OutletModel) ToDomain() *domain.Outlet {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.Outlet{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		BrandID:        m.BrandID,
		Name:           m.Name,
		Code:           m.Code,
		Address:        m.Address,
		Phone:          m.Phone,
		Latitude:       m.Latitude,
		Longitude:      m.Longitude,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromOutletDomain(d *domain.Outlet) *OutletModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &OutletModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		BrandID:        d.BrandID,
		Name:           d.Name,
		Code:           d.Code,
		Address:        d.Address,
		Phone:          d.Phone,
		Latitude:       d.Latitude,
		Longitude:      d.Longitude,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}
