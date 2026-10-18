package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email/username or password")
	ErrAccountInactive    = errors.New("user account is inactive or suspended")
	ErrInvalidPIN         = errors.New("invalid POS PIN")
	ErrTokenExpired           = errors.New("token has expired")
	ErrInvalidToken           = errors.New("invalid token")
	ErrOutletRequired         = errors.New("outlet id is required for POS login")
	ErrCurrentPasswordMismatch = errors.New("current password does not match")
	ErrCurrentPINMismatch     = errors.New("current PIN does not match")
	ErrCurrentPINRequired     = errors.New("current PIN is required")
	ErrSamePassword           = errors.New("new password cannot be the same as current password")
	ErrSamePIN                = errors.New("new PIN cannot be the same as current PIN")
	ErrTooManyAttempts        = errors.New("too many failed login attempts, account temporarily locked")
)
