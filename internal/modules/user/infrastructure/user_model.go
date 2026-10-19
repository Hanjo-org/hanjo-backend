package infrastructure

import (
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/user/domain"
)

type UserModel struct {
	ID             string         `gorm:"column:id;type:uuid;primaryKey"`
	OrganizationID string         `gorm:"column:organization_id;type:uuid;not null;index"`
	RoleID         *string        `gorm:"column:role_id;type:uuid;index"`
	BrandID        *string        `gorm:"column:brand_id;type:uuid;index"`
	OutletID       *string        `gorm:"column:outlet_id;type:uuid;index"`
	Username       string         `gorm:"column:username;type:varchar(100);not null"`
	Email          string         `gorm:"column:email;type:varchar(255);not null"`
	Password       string         `gorm:"column:password;type:varchar(255);not null"`
	Pin            *string        `gorm:"column:pin;type:varchar(255)"`
	Status         string         `gorm:"column:status;type:varchar(50);not null;default:'ACTIVE'"`
	CreatedAt      time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (UserModel) TableName() string {
	return "users"
}

func (m *UserModel) ToDomain() *domain.User {
	if m == nil {
		return nil
	}

	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		deletedAt = &m.DeletedAt.Time
	}

	return &domain.User{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		RoleID:         m.RoleID,
		BrandID:        m.BrandID,
		OutletID:       m.OutletID,
		Username:       m.Username,
		Email:          m.Email,
		Password:       m.Password,
		Pin:            m.Pin,
		Status:         m.Status,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}

func FromUserDomain(d *domain.User) *UserModel {
	if d == nil {
		return nil
	}

	var deletedAt gorm.DeletedAt
	if d.DeletedAt != nil {
		deletedAt = gorm.DeletedAt{Time: *d.DeletedAt, Valid: true}
	}

	return &UserModel{
		ID:             d.ID,
		OrganizationID: d.OrganizationID,
		RoleID:         d.RoleID,
		BrandID:        d.BrandID,
		OutletID:       d.OutletID,
		Username:       d.Username,
		Email:          d.Email,
		Password:       d.Password,
		Pin:            d.Pin,
		Status:         d.Status,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
		DeletedAt:      deletedAt,
	}
}
