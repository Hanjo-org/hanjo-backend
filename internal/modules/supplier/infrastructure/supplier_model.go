package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
)

// SupplierModel maps to the suppliers table in PostgreSQL.
type SupplierModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID string         `gorm:"column:organization_id;type:uuid;not null;index"`
	Name           string         `gorm:"column:name;type:varchar(150);not null;index"`
	Code           *string        `gorm:"column:code;type:varchar(50)"`
	ContactPerson  *string        `gorm:"column:contact_person;type:varchar(100)"`
	Phone          *string        `gorm:"column:phone;type:varchar(30)"`
	Email          *string        `gorm:"column:email;type:varchar(100)"`
	Address        *string        `gorm:"column:address;type:text"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE';index"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// TableName overrides the default table name to "suppliers".
func (SupplierModel) TableName() string {
	return "suppliers"
}

// ToDomain maps the GORM database model to the pure domain entity.
func (m *SupplierModel) ToDomain() *domain.Supplier {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.Supplier{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		Name:           m.Name,
		Code:           m.Code,
		ContactPerson:  m.ContactPerson,
		Phone:          m.Phone,
		Email:          m.Email,
		Address:        m.Address,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

// FromSupplierDomain maps the pure domain entity to the GORM database model.
func FromSupplierDomain(d *domain.Supplier) *SupplierModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &SupplierModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		Name:           d.Name,
		Code:           d.Code,
		ContactPerson:  d.ContactPerson,
		Phone:          d.Phone,
		Email:          d.Email,
		Address:        d.Address,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}
