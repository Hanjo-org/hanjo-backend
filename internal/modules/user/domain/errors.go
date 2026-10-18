package domain

import "errors"

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrUserInactive          = errors.New("user account is inactive or suspended")
	ErrEmailAlreadyExists    = errors.New("email is already registered within this organization")
	ErrUsernameAlreadyExists = errors.New("username is already registered within this organization")
	ErrInvalidCredentials    = errors.New("invalid email/username or password")
	ErrInvalidPIN            = errors.New("invalid POS PIN")
	ErrForbidden             = errors.New("forbidden: access denied to user resource")
	ErrInvalidStatus         = errors.New("invalid user status")
	ErrInvalidHierarchy      = errors.New("brand or outlet does not belong to the target organization")
)
