package usecase

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/infrastructure"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

var roleCodeRegex = regexp.MustCompile(`^[A-Z0-9_]{2,50}$`)

// RBACUsecase orchestrates role management, permission catalogs, and in-memory authorization evaluation.
type RBACUsecase interface {
	ListRoles(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*RoleResponse, int64, error)
	GetRoleByID(ctx context.Context, id string) (*RoleResponse, error)
	CreateRole(ctx context.Context, input CreateRoleInput) (*RoleResponse, error)
	UpdateRole(ctx context.Context, id string, input UpdateRoleInput) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
	GetRolesLookup(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*RoleLookupResponse, int64, error)
	GetPermissionsLookup(ctx context.Context) ([]*PermissionGroupResponse, error)
	HasPermission(roleID, permCode string) bool
	InitCache(ctx context.Context) error
}

type rbacUsecase struct {
	roleRepo  domain.RoleRepository
	permRepo  domain.PermissionRepository
	cache     infrastructure.RBACCache
	log       *zap.Logger
}

// NewRBACUsecase creates a new RBACUsecase instance.
func NewRBACUsecase(
	roleRepo domain.RoleRepository,
	permRepo domain.PermissionRepository,
	cache infrastructure.RBACCache,
	log *zap.Logger,
) RBACUsecase {
	return &rbacUsecase{
		roleRepo: roleRepo,
		permRepo: permRepo,
		cache:    cache,
		log:      log,
	}
}

func (u *rbacUsecase) InitCache(ctx context.Context) error {
	return u.cache.Reload(ctx)
}

func (u *rbacUsecase) HasPermission(roleID, permCode string) bool {
	return u.cache.HasPermission(roleID, permCode)
}

func (u *rbacUsecase) ListRoles(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*RoleResponse, int64, error) {
	roles, total, err := u.roleRepo.List(ctx, filter, params)
	if err != nil {
		u.log.Error("failed to list roles", zap.Error(err))
		return nil, 0, err
	}

	roleIDs := make([]string, len(roles))
	for i, r := range roles {
		roleIDs[i] = r.ID
	}

	permsMap, err := u.permRepo.ListByRoleIDs(ctx, roleIDs)
	if err != nil {
		u.log.Error("failed to batch fetch permissions for roles", zap.Error(err))
		return nil, 0, err
	}

	responses := make([]*RoleResponse, len(roles))
	for i, r := range roles {
		perms := permsMap[r.ID]
		responses[i] = u.toRoleResponse(r, perms)
	}

	return responses, total, nil
}

func (u *rbacUsecase) GetRoleByID(ctx context.Context, id string) (*RoleResponse, error) {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	perms, err := u.permRepo.ListByRoleID(ctx, id)
	if err != nil {
		u.log.Error("failed to fetch role permissions", zap.Error(err), zap.String("role_id", id))
		return nil, err
	}

	return u.toRoleResponse(role, perms), nil
}

func (u *rbacUsecase) CreateRole(ctx context.Context, input CreateRoleInput) (*RoleResponse, error) {
	code := strings.ToUpper(strings.TrimSpace(input.Code))
	if !roleCodeRegex.MatchString(code) {
		return nil, domain.ErrInvalidRoleCode
	}

	// Protect reserved Developer role
	if code == "DEVELOPER" {
		return nil, domain.ErrDeveloperRoleProtected
	}

	// Check for unique role code in organization
	existing, _ := u.roleRepo.GetByCode(ctx, &input.OrganizationID, code)
	if existing != nil {
		return nil, domain.ErrRoleCodeExists
	}

	tier := domain.RoleTier(input.Tier)
	role := domain.NewCustomRole(input.OrganizationID, code, strings.TrimSpace(input.Name), input.Description, tier)

	if err := u.roleRepo.Create(ctx, role, input.PermissionIDs); err != nil {
		u.log.Error("failed to create custom role", zap.Error(err), zap.String("code", code))
		return nil, err
	}

	// Invalidate and refresh in-memory cache
	_ = u.cache.Reload(ctx)

	perms, _ := u.permRepo.ListByRoleID(ctx, role.ID)
	return u.toRoleResponse(role, perms), nil
}

func (u *rbacUsecase) UpdateRole(ctx context.Context, id string, input UpdateRoleInput) (*RoleResponse, error) {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Protect system default roles and DEVELOPER role
	if role.IsSystem || role.Code == "DEVELOPER" {
		return nil, domain.ErrSystemRoleProtected
	}

	if input.Name != nil && *input.Name != "" {
		role.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		role.Description = *input.Description
	}
	if input.Status != nil && *input.Status != "" {
		role.Status = *input.Status
	}
	role.UpdatedAt = time.Now().UTC()

	if err := u.roleRepo.Update(ctx, role, input.PermissionIDs); err != nil {
		u.log.Error("failed to update role", zap.Error(err), zap.String("role_id", id))
		return nil, err
	}

	// Invalidate and refresh in-memory cache
	_ = u.cache.Reload(ctx)

	perms, _ := u.permRepo.ListByRoleID(ctx, role.ID)
	return u.toRoleResponse(role, perms), nil
}

func (u *rbacUsecase) DeleteRole(ctx context.Context, id string) error {
	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Protect system default roles and DEVELOPER role
	if role.IsSystem || role.Code == "DEVELOPER" {
		return domain.ErrSystemRoleProtected
	}

	if err := u.roleRepo.Delete(ctx, id); err != nil {
		u.log.Error("failed to delete role", zap.Error(err), zap.String("role_id", id))
		return err
	}

	// Invalidate and refresh in-memory cache
	_ = u.cache.Reload(ctx)

	return nil
}

func (u *rbacUsecase) GetRolesLookup(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*RoleLookupResponse, int64, error) {
	lookups, total, err := u.roleRepo.GetLookup(ctx, orgID, tier, params)
	if err != nil {
		u.log.Error("failed to get roles lookup", zap.Error(err), zap.String("org_id", orgID))
		return nil, 0, err
	}

	responses := make([]*RoleLookupResponse, len(lookups))
	for i, l := range lookups {
		responses[i] = ToRoleLookupResponse(l)
	}
	return responses, total, nil
}

func (u *rbacUsecase) GetPermissionsLookup(ctx context.Context) ([]*PermissionGroupResponse, error) {
	perms, err := u.permRepo.List(ctx, domain.PermissionFilter{})
	if err != nil {
		u.log.Error("failed to list permissions for lookup", zap.Error(err))
		return nil, err
	}

	groupNames := map[string]string{
		"user":      "Karyawan & Staf",
		"rbac":      "Hak Akses & Peran",
		"org":       "Organisasi & Outlet",
		"catalog":   "Menu & Katalog",
		"pos":       "Point of Sale (POS)",
		"inventory": "Inventaris & Pengadaan",
		"finance":   "Keuangan & Laporan",
	}

	// Preserve group order
	orderedGroups := []string{"user", "rbac", "org", "catalog", "pos", "inventory", "finance"}
	groupMap := make(map[string][]*PermissionResponse)

	for _, p := range perms {
		groupMap[p.Group] = append(groupMap[p.Group], ToPermissionResponse(p))
	}

	result := make([]*PermissionGroupResponse, 0, len(orderedGroups))
	for _, grp := range orderedGroups {
		list, exists := groupMap[grp]
		if !exists {
			continue
		}
		displayName := groupNames[grp]
		if displayName == "" {
			displayName = strings.ToUpper(grp)
		}
		result = append(result, &PermissionGroupResponse{
			Group:       grp,
			GroupName:   displayName,
			Permissions: list,
		})
	}

	return result, nil
}

func (u *rbacUsecase) toRoleResponse(role *domain.Role, perms []*domain.Permission) *RoleResponse {
	permResponses := make([]*PermissionResponse, len(perms))
	for i, p := range perms {
		permResponses[i] = ToPermissionResponse(p)
	}

	return &RoleResponse{
		ID:             role.ID,
		OrganizationID: role.OrganizationID,
		Code:           role.Code,
		Name:           role.Name,
		Description:    role.Description,
		Tier:           string(role.Tier),
		IsSystem:       role.IsSystem,
		Status:         role.Status,
		Permissions:    permResponses,
		CreatedAt:      role.CreatedAt,
		UpdatedAt:      role.UpdatedAt,
	}
}
