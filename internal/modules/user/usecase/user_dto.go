package usecase

// CreateUserInput carries validated data to create a new user entity.
type CreateUserInput struct {
	OrganizationID string
	RoleID         *string
	BrandID        *string
	OutletID       *string
	Username       string
	Email          string
	Password       string
	PIN            *string
	Status         string
}

// UpdateUserInput carries fields for updating an existing user entity.
type UpdateUserInput struct {
	RoleID   *string
	BrandID  *string
	OutletID *string
	Username *string
	Email    *string
	Password *string
	PIN      *string
	Status   *string
}

// UpdateStatusInput carries status change parameter.
type UpdateStatusInput struct {
	Status string
}
