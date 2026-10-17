package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email/username or password")
	ErrAccountInactive    = errors.New("user account is inactive or suspended")
	ErrInvalidPIN         = errors.New("invalid POS PIN")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrOutletRequired     = errors.New("outlet id is required for POS login")
)
