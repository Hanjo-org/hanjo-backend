package delivery

// CreateUserRequest defines payload to create a new user/staff member.
type CreateUserRequest struct {
	BrandID  *string `json:"brand_id" validate:"omitempty,uuid4"`
	OutletID *string `json:"outlet_id" validate:"omitempty,uuid4"`
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=8"`
	PIN      *string `json:"pin" validate:"omitempty,numeric,min=4,max=8"`
	Status   string  `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

// UpdateUserRequest defines payload for updating an existing user profile or assignments.
type UpdateUserRequest struct {
	BrandID  *string `json:"brand_id" validate:"omitempty,uuid4"`
	OutletID *string `json:"outlet_id" validate:"omitempty,uuid4"`
	Username *string `json:"username" validate:"omitempty,min=3,max=50"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Password *string `json:"password" validate:"omitempty,min=8"`
	PIN      *string `json:"pin" validate:"omitempty,numeric,min=4,max=8"`
	Status   *string `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE SUSPENDED"`
}

// UpdateUserStatusRequest defines payload for toggling user active status.
type UpdateUserStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE SUSPENDED"`
}
