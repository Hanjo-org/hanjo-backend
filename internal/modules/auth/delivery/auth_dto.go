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

// ChangePasswordRequest defines body payload for updating own password.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

// ChangePINRequest defines body payload for updating or setting own POS PIN.
type ChangePINRequest struct {
	CurrentPIN *string `json:"current_pin" validate:"omitempty,numeric,min=4,max=8"`
	NewPIN     string  `json:"new_pin" validate:"required,numeric,min=4,max=8"`
}
