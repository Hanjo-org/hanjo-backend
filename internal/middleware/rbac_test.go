package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	rbacdomain "github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	rbacusecase "github.com/hanjo-erp/hanjo/internal/modules/rbac/usecase"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Mock RBAC Usecase
type mockRBACUsecase struct {
	hasPermissionFn func(roleID, permCode string) bool
}

func (m *mockRBACUsecase) ListRoles(ctx context.Context, filter rbacdomain.RoleFilter, params response.PaginationParams) ([]*rbacusecase.RoleResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockRBACUsecase) GetRoleByID(ctx context.Context, id string) (*rbacusecase.RoleResponse, error) {
	return nil, nil
}
func (m *mockRBACUsecase) CreateRole(ctx context.Context, input rbacusecase.CreateRoleInput) (*rbacusecase.RoleResponse, error) {
	return nil, nil
}
func (m *mockRBACUsecase) UpdateRole(ctx context.Context, id string, input rbacusecase.UpdateRoleInput) (*rbacusecase.RoleResponse, error) {
	return nil, nil
}
func (m *mockRBACUsecase) DeleteRole(ctx context.Context, id string) error {
	return nil
}
func (m *mockRBACUsecase) GetRolesLookup(ctx context.Context, orgID string, tier *rbacdomain.RoleTier, params response.PaginationParams) ([]*rbacusecase.RoleLookupResponse, int64, error) {
	return nil, 0, nil
}
func (m *mockRBACUsecase) GetPermissionsLookup(ctx context.Context) ([]*rbacusecase.PermissionGroupResponse, error) {
	return nil, nil
}
func (m *mockRBACUsecase) HasPermission(roleID, permCode string) bool {
	if m.hasPermissionFn != nil {
		return m.hasPermissionFn(roleID, permCode)
	}
	return false
}
func (m *mockRBACUsecase) InitCache(ctx context.Context) error {
	return nil
}

// Mock User Repo for PIN verification
type mockUserRepoForPIN struct {
	getByOutletAndUsernameFn func(ctx context.Context, outletID, username string) (*userdomain.User, error)
	getByUsernameFn          func(ctx context.Context, orgID, username string) (*userdomain.User, error)
	listFn                   func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error)
}

func (m *mockUserRepoForPIN) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	return nil, userdomain.ErrUserNotFound
}
func (m *mockUserRepoForPIN) GetByIDs(ctx context.Context, ids []string) ([]*userdomain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForPIN) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	return nil, userdomain.ErrUserNotFound
}
func (m *mockUserRepoForPIN) GetByUsername(ctx context.Context, orgID, username string) (*userdomain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, orgID, username)
	}
	return nil, userdomain.ErrUserNotFound
}
func (m *mockUserRepoForPIN) GetByOutletAndUsername(ctx context.Context, outletID, username string) (*userdomain.User, error) {
	if m.getByOutletAndUsernameFn != nil {
		return m.getByOutletAndUsernameFn(ctx, outletID, username)
	}
	return nil, userdomain.ErrUserNotFound
}
func (m *mockUserRepoForPIN) List(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, params)
	}
	return nil, 0, nil
}
func (m *mockUserRepoForPIN) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForPIN) ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForPIN) GetLookup(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.UserLookup, int64, error) {
	return nil, 0, nil
}
func (m *mockUserRepoForPIN) Create(ctx context.Context, user *userdomain.User) error {
	return nil
}
func (m *mockUserRepoForPIN) Update(ctx context.Context, user *userdomain.User) error {
	return nil
}
func (m *mockUserRepoForPIN) Delete(ctx context.Context, id string) error {
	return nil
}

func TestRequirePermission(t *testing.T) {
	roleID := uuid.NewString()

	uc := &mockRBACUsecase{
		hasPermissionFn: func(rID, perm string) bool {
			return rID == roleID && perm == "catalog:product:create"
		},
	}

	app := fiber.New()
	app.Get("/products/create",
		func(c fiber.Ctx) error {
			// Simulate injected context
			if role := c.Get("X-Mock-Role"); role != "" {
				contextutil.SetUserRole(c, role)
			}
			if rID := c.Get("X-Mock-Role-ID"); rID != "" {
				contextutil.SetUserRoleID(c, rID)
			}
			return c.Next()
		},
		middleware.RequirePermission(uc, "catalog:product:create"),
		func(c fiber.Ctx) error {
			return c.SendString("allowed")
		},
	)

	t.Run("developer role bypasses check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/create", nil)
		req.Header.Set("X-Mock-Role", "DEVELOPER")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("role has permission", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/create", nil)
		req.Header.Set("X-Mock-Role-ID", roleID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("role lacks permission", func(t *testing.T) {
		otherRoleID := uuid.NewString()
		req := httptest.NewRequest(http.MethodGet, "/products/create", nil)
		req.Header.Set("X-Mock-Role-ID", otherRoleID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("missing role id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/products/create", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestRequireAnyPermission(t *testing.T) {
	roleID := uuid.NewString()

	uc := &mockRBACUsecase{
		hasPermissionFn: func(rID, perm string) bool {
			return rID == roleID && (perm == "order:read" || perm == "order:update")
		},
	}

	app := fiber.New()
	app.Get("/orders",
		func(c fiber.Ctx) error {
			if role := c.Get("X-Mock-Role"); role != "" {
				contextutil.SetUserRole(c, role)
			}
			if rID := c.Get("X-Mock-Role-ID"); rID != "" {
				contextutil.SetUserRoleID(c, rID)
			}
			return c.Next()
		},
		middleware.RequireAnyPermission(uc, "order:read", "order:manage"),
		func(c fiber.Ctx) error {
			return c.SendString("allowed")
		},
	)

	t.Run("developer bypass", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("X-Mock-Role", "DEVELOPER")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("has one of required permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("X-Mock-Role-ID", roleID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("has none of required permissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("X-Mock-Role-ID", uuid.NewString())
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

func TestRequirePermissionOrSupervisorPIN(t *testing.T) {
	orgID := uuid.NewString()
	outletID := uuid.NewString()
	cashierRoleID := uuid.NewString()
	supervisorRoleID := uuid.NewString()

	supervisorHashedPIN, err := password.Hash("998877")
	require.NoError(t, err)

	supervisorUser := &userdomain.User{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		OutletID:       &outletID,
		Username:       "manager.spv",
		RoleID:         &supervisorRoleID,
		Status:         "ACTIVE",
		Pin:            ptr.Ptr(supervisorHashedPIN),
	}

	uc := &mockRBACUsecase{
		hasPermissionFn: func(rID, perm string) bool {
			// Cashier lacks void permission, supervisor has it
			return rID == supervisorRoleID && perm == "pos:order:void"
		},
	}

	userRepo := &mockUserRepoForPIN{
		getByOutletAndUsernameFn: func(ctx context.Context, outID, username string) (*userdomain.User, error) {
			if outID == outletID && username == "manager.spv" {
				return supervisorUser, nil
			}
			return nil, userdomain.ErrUserNotFound
		},
		listFn: func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error) {
			return []*userdomain.User{supervisorUser}, 1, nil
		},
	}

	app := fiber.New()
	app.Post("/orders/void",
		func(c fiber.Ctx) error {
			if role := c.Get("X-Mock-Role"); role != "" {
				contextutil.SetUserRole(c, role)
			}
			if rID := c.Get("X-Mock-Role-ID"); rID != "" {
				contextutil.SetUserRoleID(c, rID)
			}
			contextutil.SetOrganizationID(c, orgID)
			contextutil.SetOutletID(c, outletID)
			return c.Next()
		},
		middleware.RequirePermissionOrSupervisorPIN(uc, userRepo, "pos:order:void"),
		func(c fiber.Ctx) error {
			spvID := contextutil.GetSupervisorID(c)
			return c.SendString("voided, supervisor=" + spvID)
		},
	)

	t.Run("supervisor directly authorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/void", nil)
		req.Header.Set("X-Mock-Role-ID", supervisorRoleID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cashier without supervisor PIN rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/void", nil)
		req.Header.Set("X-Mock-Role-ID", cashierRoleID)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("cashier with valid supervisor username and PIN approved", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/void", nil)
		req.Header.Set("X-Mock-Role-ID", cashierRoleID)
		req.Header.Set("X-Supervisor-Username", "manager.spv")
		req.Header.Set("X-Supervisor-PIN", "998877")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("cashier with invalid supervisor PIN rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/void", nil)
		req.Header.Set("X-Mock-Role-ID", cashierRoleID)
		req.Header.Set("X-Supervisor-Username", "manager.spv")
		req.Header.Set("X-Supervisor-PIN", "111111")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("cashier with valid PIN without username searches outlet approved", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/orders/void", nil)
		req.Header.Set("X-Mock-Role-ID", cashierRoleID)
		req.Header.Set("X-Supervisor-PIN", "998877")
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
