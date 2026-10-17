package usecase

import (
	"time"

	"github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
)

type LoginInput struct {
	Identifier string
	Password   string
}

type POSLoginInput struct {
	OutletID string
	Username string
	PIN      string
}

type UserOutput struct {
	ID             string  `json:"id"`
	OrganizationID string  `json:"organization_id"`
	BrandID        *string `json:"brand_id,omitempty"`
	OutletID       *string `json:"outlet_id,omitempty"`
	Username       string  `json:"username"`
	Email          string  `json:"email"`
	Status         string  `json:"status"`
}

type AuthOutput struct {
	User   UserOutput       `json:"user"`
	Tokens domain.TokenPair `json:"tokens"`
}

type UserProfileOutput struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	BrandID        *string    `json:"brand_id,omitempty"`
	OutletID       *string    `json:"outlet_id,omitempty"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
