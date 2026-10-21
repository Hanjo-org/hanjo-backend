package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
)

// UoMModel maps to the uoms table in PostgreSQL.
type UoMModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID *string        `gorm:"column:organization_id;type:uuid;index"`
	Name           string         `gorm:"column:name;type:varchar(100);not null"`
	Symbol         string         `gorm:"column:symbol;type:varchar(20);not null"`
	Category       string         `gorm:"column:category;type:varchar(50);not null;index"`
	DecimalPlaces  int            `gorm:"column:decimal_places;type:smallint;not null;default:0"`
	IsSystem       bool           `gorm:"column:is_system;not null;default:false"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE';index"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UoMModel) TableName() string {
	return "uoms"
}

func (m *UoMModel) ToDomain() *domain.UoM {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.UoM{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		Name:           m.Name,
		Symbol:         m.Symbol,
		Category:       domain.UoMCategory(m.Category),
		DecimalPlaces:  m.DecimalPlaces,
		IsSystem:       m.IsSystem,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromUoMDomain(d *domain.UoM) *UoMModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &UoMModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		Name:           d.Name,
		Symbol:         d.Symbol,
		Category:       string(d.Category),
		DecimalPlaces:  d.DecimalPlaces,
		IsSystem:       d.IsSystem,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

// UoMConversionModel maps to the uom_conversions table in PostgreSQL.
type UoMConversionModel struct {
	ID             string    `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID *string   `gorm:"column:organization_id;type:uuid;index"`
	FromUoMID      string    `gorm:"column:from_uom_id;type:uuid;not null;index"`
	ToUoMID        string    `gorm:"column:to_uom_id;type:uuid;not null;index"`
	Factor         float64   `gorm:"column:factor;type:numeric(15,6);not null"`
	IsSystem       bool      `gorm:"column:is_system;not null;default:false"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

func (UoMConversionModel) TableName() string {
	return "uom_conversions"
}

func (m *UoMConversionModel) ToDomain() *domain.UoMConversion {
	if m == nil {
		return nil
	}

	return &domain.UoMConversion{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		FromUoMID:      m.FromUoMID,
		ToUoMID:        m.ToUoMID,
		Factor:         m.Factor,
		IsSystem:       m.IsSystem,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
}

func FromConversionDomain(d *domain.UoMConversion) *UoMConversionModel {
	if d == nil {
		return nil
	}

	return &UoMConversionModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		FromUoMID:      d.FromUoMID,
		ToUoMID:        d.ToUoMID,
		Factor:         d.Factor,
		IsSystem:       d.IsSystem,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}
