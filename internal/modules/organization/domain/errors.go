package domain

import "errors"

var (
	ErrOrganizationNotFound    = errors.New("organization not found")
	ErrOrganizationSlugExists  = errors.New("organization slug already exists")
	ErrBrandNotFound           = errors.New("brand not found")
	ErrBrandSlugExists         = errors.New("brand slug already exists within organization")
	ErrOutletNotFound          = errors.New("outlet not found")
	ErrOutletCodeExists        = errors.New("outlet code already exists within organization")
	ErrInvalidStatus           = errors.New("invalid status value")
)
