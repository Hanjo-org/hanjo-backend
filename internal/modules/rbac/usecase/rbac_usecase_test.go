package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/infrastructure"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/usecase"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockRoleRepository struct {
	getByIDFn                       func(ctx context.Context, id string) (*domain.Role, error)
	getByIDsFn                      func(ctx context.Context, ids []string) ([]*domain.Role, error)
	getByCodeFn                     func(ctx context.Context, orgID *string, code string) (*domain.Role, error)
	listFn                          func(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*domain.Role, int64, error)
	getLookupFn                     func(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*domain.RoleLookup, int64, error)
	createFn                        func(ctx context.Context, role *domain.Role, permissionIDs []string) error
	updateFn                        func(ctx context.Context, role *domain.Role, permissionIDs []string) error
	deleteFn                        func(ctx context.Context, id string) error
	getAllRolePermissionsMappingFn func(ctx context.Context) (map[string][]string, error)
	getAllRoleCodesFn               func(ctx context.Context) (map[string]string, error)
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Role, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockRoleRepository) GetByCode(ctx context.Context, orgID *string, code string) (*domain.Role, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, orgID, code)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRoleRepository) List(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*domain.Role, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, params)
	}
	return []*domain.Role{}, 0, nil
}

func (m *mockRoleRepository) GetLookup(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*domain.RoleLookup, int64, error) {
	if m.getLookupFn != nil {
		return m.getLookupFn(ctx, orgID, tier, params)
	}
	return []*domain.RoleLookup{}, 0, nil
}

func (m *mockRoleRepository) Create(ctx context.Context, role *domain.Role, permissionIDs []string) error {
	if m.createFn != nil {
		return m.createFn(ctx, role, permissionIDs)
	}
	return nil
}

func (m *mockRoleRepository) Update(ctx context.Context, role *domain.Role, permissionIDs []string) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, role, permissionIDs)
	}
	return nil
}

func (m *mockRoleRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRoleRepository) GetAllRolePermissionsMapping(ctx context.Context) (map[string][]string, error) {
	if m.getAllRolePermissionsMappingFn != nil {
		return m.getAllRolePermissionsMappingFn(ctx)
	}
	return make(map[string][]string), nil
}

func (m *mockRoleRepository) GetAllRoleCodes(ctx context.Context) (map[string]string, error) {
	if m.getAllRoleCodesFn != nil {
		return m.getAllRoleCodesFn(ctx)
	}
	return make(map[string]string), nil
}

type mockPermissionRepository struct {
	getByIDFn        func(ctx context.Context, id string) (*domain.Permission, error)
	getByCodeFn      func(ctx context.Context, code string) (*domain.Permission, error)
	getByCodesFn     func(ctx context.Context, codes []string) ([]*domain.Permission, error)
	listFn           func(ctx context.Context, filter domain.PermissionFilter) ([]*domain.Permission, error)
	listByRoleIDFn   func(ctx context.Context, roleID string) ([]*domain.Permission, error)
	listByRoleIDsFn  func(ctx context.Context, roleIDs []string) (map[string][]*domain.Permission, error)
}

func (m *mockPermissionRepository) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrPermissionNotFound
}

func (m *mockPermissionRepository) GetByCode(ctx context.Context, code string) (*domain.Permission, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, code)
	}
	return nil, domain.ErrPermissionNotFound
}

func (m *mockPermissionRepository) GetByCodes(ctx context.Context, codes []string) ([]*domain.Permission, error) {
	if m.getByCodesFn != nil {
		return m.getByCodesFn(ctx, codes)
	}
	return []*domain.Permission{}, nil
}

func (m *mockPermissionRepository) List(ctx context.Context, filter domain.PermissionFilter) ([]*domain.Permission, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter)
	}
	return []*domain.Permission{}, nil
}

func (m *mockPermissionRepository) ListByRoleID(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	if m.listByRoleIDFn != nil {
		return m.listByRoleIDFn(ctx, roleID)
	}
	return []*domain.Permission{}, nil
}

func (m *mockPermissionRepository) ListByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]*domain.Permission, error) {
	if m.listByRoleIDsFn != nil {
		return m.listByRoleIDsFn(ctx, roleIDs)
	}
	return make(map[string][]*domain.Permission), nil
}

func TestRBACUsecase_CreateRole(t *testing.T) {
	logger := zap.NewNop()

	t.Run("success create custom role", func(t *testing.T) {
		roleRepo := &mockRoleRepository{
			getByCodeFn: func(ctx context.Context, orgID *string, code string) (*domain.Role, error) {
				return nil, domain.ErrRoleNotFound
			},
			createFn: func(ctx context.Context, role *domain.Role, permissionIDs []string) error {
				assert.Equal(t, "HEAD_BARISTA", role.Code)
				assert.Equal(t, domain.RoleTierOutlet, role.Tier)
				return nil
			},
		}
		permRepo := &mockPermissionRepository{
			listByRoleIDFn: func(ctx context.Context, roleID string) ([]*domain.Permission, error) {
				return []*domain.Permission{
					{ID: "p-1", Code: "pos:order:create", Group: "pos", Name: "Buat Pesanan"},
				}, nil
			},
		}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		res, err := uc.CreateRole(context.Background(), usecase.CreateRoleInput{
			OrganizationID: "org-1",
			Code:           "head_barista", // test auto-uppercase
			Name:           "Head Barista",
			Description:    "Supervisor Barista Cabang",
			Tier:           "OUTLET",
			PermissionIDs:  []string{"p-1"},
		})

		require.NoError(t, err)
		assert.Equal(t, "HEAD_BARISTA", res.Code)
		assert.Equal(t, "Head Barista", res.Name)
		assert.Len(t, res.Permissions, 1)
	})

	t.Run("reject reserved DEVELOPER role code", func(t *testing.T) {
		roleRepo := &mockRoleRepository{}
		permRepo := &mockPermissionRepository{}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		_, err := uc.CreateRole(context.Background(), usecase.CreateRoleInput{
			OrganizationID: "org-1",
			Code:           "DEVELOPER",
			Name:           "Dev Role",
			Tier:           "ORGANIZATION",
		})

		assert.ErrorIs(t, err, domain.ErrDeveloperRoleProtected)
	})

	t.Run("reject duplicate role code in organization", func(t *testing.T) {
		roleRepo := &mockRoleRepository{
			getByCodeFn: func(ctx context.Context, orgID *string, code string) (*domain.Role, error) {
				return &domain.Role{ID: "r-1", Code: "CASHIER"}, nil
			},
		}
		permRepo := &mockPermissionRepository{}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		_, err := uc.CreateRole(context.Background(), usecase.CreateRoleInput{
			OrganizationID: "org-1",
			Code:           "CASHIER",
			Name:           "Kasir",
			Tier:           "OUTLET",
		})

		assert.ErrorIs(t, err, domain.ErrRoleCodeExists)
	})
}

func TestRBACUsecase_UpdateAndDeleteProtection(t *testing.T) {
	logger := zap.NewNop()

	t.Run("reject update on system role", func(t *testing.T) {
		roleRepo := &mockRoleRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Role, error) {
				return &domain.Role{
					ID:       id,
					Code:     "CASHIER",
					IsSystem: true,
				}, nil
			},
		}
		permRepo := &mockPermissionRepository{}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		_, err := uc.UpdateRole(context.Background(), "sys-role-1", usecase.UpdateRoleInput{
			Name: ptr.Ptr("New Cashier Name"),
		})

		assert.ErrorIs(t, err, domain.ErrSystemRoleProtected)
	})

	t.Run("reject delete on system role", func(t *testing.T) {
		roleRepo := &mockRoleRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Role, error) {
				return &domain.Role{
					ID:       id,
					Code:     "OWNER",
					IsSystem: true,
				}, nil
			},
		}
		permRepo := &mockPermissionRepository{}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		err := uc.DeleteRole(context.Background(), "sys-role-2")
		assert.ErrorIs(t, err, domain.ErrSystemRoleProtected)
	})
}

func TestRBACUsecase_Lookups(t *testing.T) {
	logger := zap.NewNop()

	t.Run("get roles lookup excludes DEVELOPER", func(t *testing.T) {
		roleRepo := &mockRoleRepository{
			getLookupFn: func(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*domain.RoleLookup, int64, error) {
				return []*domain.RoleLookup{
					{ID: "r-1", Code: "OWNER", Name: "Owner", Tier: domain.RoleTierOrganization},
					{ID: "r-2", Code: "CASHIER", Name: "Kasir POS", Tier: domain.RoleTierOutlet},
				}, 2, nil
			},
		}
		permRepo := &mockPermissionRepository{}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		res, total, err := uc.GetRolesLookup(context.Background(), "org-1", nil, response.PaginationParams{Page: 1, PerPage: 10})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, res, 2)
		for _, r := range res {
			assert.NotEqual(t, "DEVELOPER", r.Code)
		}
	})

	t.Run("get permissions lookup grouped by domain", func(t *testing.T) {
		roleRepo := &mockRoleRepository{}
		permRepo := &mockPermissionRepository{
			listFn: func(ctx context.Context, filter domain.PermissionFilter) ([]*domain.Permission, error) {
				return []*domain.Permission{
					{ID: "p-1", Code: "user:staff:read", Group: "user", Name: "Lihat Staf"},
					{ID: "p-2", Code: "pos:order:create", Group: "pos", Name: "Buat Pesanan"},
					{ID: "p-3", Code: "pos:order:void", Group: "pos", Name: "Void Pesanan"},
				}, nil
			},
		}
		cache := infrastructure.NewRBACCache(roleRepo, logger)
		uc := usecase.NewRBACUsecase(roleRepo, permRepo, cache, logger)

		groups, err := uc.GetPermissionsLookup(context.Background())
		require.NoError(t, err)
		assert.Len(t, groups, 2) // user & pos groups
		assert.Equal(t, "user", groups[0].Group)
		assert.Equal(t, "pos", groups[1].Group)
		assert.Len(t, groups[1].Permissions, 2)
	})
}

func TestRBACCache_DeveloperBypassAndEvaluation(t *testing.T) {
	logger := zap.NewNop()

	roleRepo := &mockRoleRepository{
		getAllRolePermissionsMappingFn: func(ctx context.Context) (map[string][]string, error) {
			return map[string][]string{
				"role-cashier": {"pos:order:create", "pos:order:payment"},
			}, nil
		},
		getAllRoleCodesFn: func(ctx context.Context) (map[string]string, error) {
			return map[string]string{
				"role-dev":     "DEVELOPER",
				"role-cashier": "CASHIER",
			}, nil
		},
	}

	cache := infrastructure.NewRBACCache(roleRepo, logger)
	err := cache.Reload(context.Background())
	require.NoError(t, err)

	// 1. DEVELOPER role has full bypass: any permission returns true
	assert.True(t, cache.HasPermission("role-dev", "pos:order:create"))
	assert.True(t, cache.HasPermission("role-dev", "pos:order:void"))
	assert.True(t, cache.HasPermission("role-dev", "non_existent_crazy_permission"))

	// 2. CASHIER role has only assigned permissions
	assert.True(t, cache.HasPermission("role-cashier", "pos:order:create"))
	assert.True(t, cache.HasPermission("role-cashier", "pos:order:payment"))
	assert.False(t, cache.HasPermission("role-cashier", "pos:order:void"))
	assert.False(t, cache.HasPermission("role-cashier", "inventory:stock:read"))

	// 3. Unknown role returns false
	assert.False(t, cache.HasPermission("unknown-role", "pos:order:create"))
	assert.False(t, cache.HasPermission("", "pos:order:create"))
}
