package domain

import "errors"

// Sentinel errors for the Supplier domain
var (
	ErrSupplierNotFound      = errors.New("supplier not found")
	ErrDuplicateSupplierCode = errors.New("supplier code already exists")
	ErrInvalidSupplierStatus = errors.New("invalid supplier status")
)
