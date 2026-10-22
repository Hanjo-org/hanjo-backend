package usecase

import (
	"time"
)

// CreateSupplierRequest represents the payload for creating a new supplier.
type CreateSupplierRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=150"`
	Code          *string `json:"code,omitempty" validate:"omitempty,min=1,max=50"`
	ContactPerson *string `json:"contact_person,omitempty" validate:"omitempty,max=100"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,max=30"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=100"`
	Address       *string `json:"address,omitempty" validate:"omitempty,max=500"`
}

// UpdateSupplierRequest represents the payload for updating an existing supplier.
type UpdateSupplierRequest struct {
	Name          string  `json:"name" validate:"required,min=2,max=150"`
	Code          *string `json:"code,omitempty" validate:"omitempty,min=1,max=50"`
	ContactPerson *string `json:"contact_person,omitempty" validate:"omitempty,max=100"`
	Phone         *string `json:"phone,omitempty" validate:"omitempty,max=30"`
	Email         *string `json:"email,omitempty" validate:"omitempty,email,max=100"`
	Address       *string `json:"address,omitempty" validate:"omitempty,max=500"`
	Status        string  `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}

// SupplierResponse represents the full supplier details for backoffice administration.
type SupplierResponse struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	Code           *string   `json:"code,omitempty"`
	ContactPerson  *string   `json:"contact_person,omitempty"`
	Phone          *string   `json:"phone,omitempty"`
	Email          *string   `json:"email,omitempty"`
	Address        *string   `json:"address,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// SupplierLookupResponse represents a lightweight, non-sensitive item for UI comboboxes.
type SupplierLookupResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Code          *string `json:"code,omitempty"`
	ContactPerson *string `json:"contact_person,omitempty"`
	Phone         *string `json:"phone,omitempty"`
}
