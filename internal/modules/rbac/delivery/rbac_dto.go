package delivery

// CreateRoleRequest defines payload for creating a custom organization role.
type CreateRoleRequest struct {
	Code          string   `json:"code" validate:"required,min=2,max=50"`
	Name          string   `json:"name" validate:"required,min=2,max=100"`
	Description   string   `json:"description" validate:"max=500"`
	Tier          string   `json:"tier" validate:"required,oneof=ORGANIZATION BRAND OUTLET"`
	PermissionIDs []string `json:"permission_ids" validate:"omitempty"`
}

// UpdateRoleRequest defines payload for modifying an existing custom role.
type UpdateRoleRequest struct {
	Name          *string  `json:"name" validate:"omitempty,min=2,max=100"`
	Description   *string  `json:"description" validate:"omitempty,max=500"`
	PermissionIDs []string `json:"permission_ids" validate:"omitempty"`
	Status        *string  `json:"status" validate:"omitempty,oneof=ACTIVE INACTIVE"`
}
