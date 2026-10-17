package delivery

// LoginRequest defines body payload for standard dashboard login.
type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required,min=6"`
}

// POSLoginRequest defines body payload for fast POS PIN login at physical outlet.
type POSLoginRequest struct {
	OutletID string `json:"outlet_id" validate:"required,uuid4"`
	Username string `json:"username" validate:"required"`
	PIN      string `json:"pin" validate:"required,numeric,min=4,max=8"`
}

// RefreshTokenRequest defines optional body payload when refresh token is not sent via cookie.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
