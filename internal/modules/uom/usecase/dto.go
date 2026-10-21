package usecase

import (
	"time"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
)

// CreateUoMRequest represents the payload for creating a custom UoM.
type CreateUoMRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=100"`
	Symbol        string `json:"symbol" validate:"required,min=1,max=20"`
	Category      string `json:"category" validate:"required,oneof=WEIGHT VOLUME UNIT"`
	DecimalPlaces int    `json:"decimal_places" validate:"min=0,max=6"`
}

// UpdateUoMRequest represents the payload for updating an existing custom UoM.
type UpdateUoMRequest struct {
	Name   string `json:"name" validate:"required,min=2,max=100"`
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}

// CreateConversionRequest represents the payload for creating a unit conversion rule.
type CreateConversionRequest struct {
	FromUoMID string  `json:"from_uom_id" validate:"required,uuid"`
	ToUoMID   string  `json:"to_uom_id" validate:"required,uuid"`
	Factor    float64 `json:"factor" validate:"required,gt=0"`
}

// ConvertQuantityRequest defines the payload for runtime unit conversion calculations.
type ConvertQuantityRequest struct {
	FromUoMID string  `json:"from_uom_id" validate:"required,uuid"`
	ToUoMID   string  `json:"to_uom_id" validate:"required,uuid"`
	Quantity  float64 `json:"quantity" validate:"required,gt=0"`
}

// ConvertQuantityResponse returns the result of a runtime unit conversion.
type ConvertQuantityResponse struct {
	FromUoM           domain.UoMSummary `json:"from_uom"`
	ToUoM             domain.UoMSummary `json:"to_uom"`
	OriginalQuantity  float64           `json:"original_quantity"`
	ConvertedQuantity float64           `json:"converted_quantity"`
	Factor            float64           `json:"factor"`
}

// UoMResponse represents full UoM details for backoffice administration.
type UoMResponse struct {
	ID             string             `json:"id"`
	OrganizationID *string            `json:"organization_id,omitempty"`
	Name           string             `json:"name"`
	Symbol         string             `json:"symbol"`
	Category       domain.UoMCategory `json:"category"`
	DecimalPlaces  int                `json:"decimal_places"`
	IsSystem       bool               `json:"is_system"`
	Status         string             `json:"status"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// UoMLookupResponse represents a lightweight, non-sensitive item for UI comboboxes.
type UoMLookupResponse struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Symbol        string             `json:"symbol"`
	Category      domain.UoMCategory `json:"category"`
	DecimalPlaces int                `json:"decimal_places"`
}

// UoMConversionResponse represents a unit conversion with resolved unit names.
type UoMConversionResponse struct {
	ID             string    `json:"id"`
	OrganizationID *string   `json:"organization_id,omitempty"`
	FromUoMID      string    `json:"from_uom_id"`
	FromUoMName    string    `json:"from_uom_name"`
	FromUoMSymbol  string    `json:"from_uom_symbol"`
	ToUoMID        string    `json:"to_uom_id"`
	ToUoMName      string    `json:"to_uom_name"`
	ToUoMSymbol    string    `json:"to_uom_symbol"`
	Factor         float64   `json:"factor"`
	IsSystem       bool      `json:"is_system"`
	CreatedAt      time.Time `json:"created_at"`
}
