package domain

import "errors"

// Sentinel errors for the UoM domain
var (
	ErrUoMNotFound              = errors.New("uom not found")
	ErrDuplicateSymbol          = errors.New("uom symbol already exists")
	ErrSystemUoMImmutable       = errors.New("system default uom cannot be modified or deleted")
	ErrInvalidUoMCategory       = errors.New("invalid uom category")
	ErrConversionNotFound       = errors.New("uom conversion not found")
	ErrDuplicateConversion      = errors.New("conversion already exists between these units")
	ErrInvalidConversionFactor  = errors.New("conversion factor must be greater than zero")
	ErrIdenticalConversionUnits = errors.New("from and to uom must be different")
	ErrCategoryMismatch         = errors.New("conversion can only be established between units of the same category")
)
