package domain

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Supplier represents a vendor/supplier master entity.
type Supplier struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	Name           string     `json:"name"`
	Code           *string    `json:"code,omitempty"`
	ContactPerson  *string    `json:"contact_person,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Address        *string    `json:"address,omitempty"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
}

// NewSupplier creates a new Supplier entity.
func NewSupplier(orgID, name string, code, contactPerson, phone, email, address *string) *Supplier {
	now := time.Now().UTC()
	return &Supplier{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		Name:           name,
		Code:           code,
		ContactPerson:  contactPerson,
		Phone:          phone,
		Email:          email,
		Address:        address,
		Status:         "ACTIVE",
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// SupplierSummary provides a lightweight summary for foreign domain cross-referencing.
type SupplierSummary struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Code          *string `json:"code,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Phone         *string `json:"phone,omitempty"`
	Status        string  `json:"status"`
}

// ToSummary converts the Supplier entity into a foreign-friendly summary.
func (s *Supplier) ToSummary() SupplierSummary {
	return SupplierSummary{
		ID:            s.ID,
		Name:          s.Name,
		Code:          s.Code,
		ContactPerson: s.ContactPerson,
		Phone:         s.Phone,
		Status:        s.Status,
	}
}

// SupplierFilter specifies filtering criteria for querying suppliers.
type SupplierFilter struct {
	Search string
	Status *string
}

// SupplierRepository defines the persistence contract for Supplier entities.
type SupplierRepository interface {
	GetByID(ctx context.Context, id string) (*Supplier, error)
	GetByIDs(ctx context.Context, ids []string) ([]*Supplier, error)
	GetByCode(ctx context.Context, orgID, code string) (*Supplier, error)
	List(ctx context.Context, orgID string, filter SupplierFilter, pagination response.PaginationParams) ([]*Supplier, int64, error)
	ListLookup(ctx context.Context, orgID string) ([]*Supplier, error)
	Create(ctx context.Context, entity *Supplier) error
	Update(ctx context.Context, entity *Supplier) error
	Delete(ctx context.Context, id string) error
}
