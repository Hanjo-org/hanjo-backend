package delivery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/delivery"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/usecase"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockRBACUsecase struct {
	listRolesFn            func(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*usecase.RoleResponse, int64, error)
	getRoleByIDFn          func(ctx context.Context, id string) (*usecase.RoleResponse, error)
	createRoleFn           func(ctx context.Context, input usecase.CreateRoleInput) (*usecase.RoleResponse, error)
	updateRoleFn           func(ctx context.Context, id string, input usecase.UpdateRoleInput) (*usecase.RoleResponse, error)
	deleteRoleFn           func(ctx context.Context, id string) error
	getRolesLookupFn       func(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*usecase.RoleLookupResponse, int64, error)
	getPermissionsLookupFn func(ctx context.Context) ([]*usecase.PermissionGroupResponse, error)
	hasPermissionFn        func(roleID, permCode string) bool
	initCacheFn            func(ctx context.Context) error
}

func (m *mockRBACUsecase) ListRoles(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*usecase.RoleResponse, int64, error) {
	if m.listRolesFn != nil {
		return m.listRolesFn(ctx, filter, params)
	}
	return []*usecase.RoleResponse{}, 0, nil
}

func (m *mockRBACUsecase) GetRoleByID(ctx context.Context, id string) (*usecase.RoleResponse, error) {
	if m.getRoleByIDFn != nil {
		return m.getRoleByIDFn(ctx, id)
	}
	return nil, domain.ErrRoleNotFound
}

func (m *mockRBACUsecase) CreateRole(ctx context.Context, input usecase.CreateRoleInput) (*usecase.RoleResponse, error) {
	if m.createRoleFn != nil {
		return m.createRoleFn(ctx, input)
	}
	return nil, nil
}

func (m *mockRBACUsecase) UpdateRole(ctx context.Context, id string, input usecase.UpdateRoleInput) (*usecase.RoleResponse, error) {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, id, input)
	}
	return nil, nil
}

func (m *mockRBACUsecase) DeleteRole(ctx context.Context, id string) error {
	if m.deleteRoleFn != nil {
		return m.deleteRoleFn(ctx, id)
	}
	return nil
}

func (m *mockRBACUsecase) GetRolesLookup(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*usecase.RoleLookupResponse, int64, error) {
	if m.getRolesLookupFn != nil {
		return m.getRolesLookupFn(ctx, orgID, tier, params)
	}
	return []*usecase.RoleLookupResponse{}, 0, nil
}

func (m *mockRBACUsecase) GetPermissionsLookup(ctx context.Context) ([]*usecase.PermissionGroupResponse, error) {
	if m.getPermissionsLookupFn != nil {
		return m.getPermissionsLookupFn(ctx)
	}
	return []*usecase.PermissionGroupResponse{}, nil
}

func (m *mockRBACUsecase) HasPermission(roleID, permCode string) bool {
	if m.hasPermissionFn != nil {
		return m.hasPermissionFn(roleID, permCode)
	}
	return false
}

func (m *mockRBACUsecase) InitCache(ctx context.Context) error {
	if m.initCacheFn != nil {
		return m.initCacheFn(ctx)
	}
	return nil
}

func setupTestApp(t *testing.T, uc usecase.RBACUsecase) (*fiber.App, *jwt.Manager, string) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtManager := jwt.NewRSAManager(privKey, pubKey, "hanjo-test", "hanjo-client")
	token, err := jwtManager.GenerateToken(jwt.TokenPayload{
		UserID:         uuid.NewString(),
		Role:           "OWNER",
		OrganizationID: uuid.NewString(),
		Duration:       time.Hour,
	})
	require.NoError(t, err)

	app := fiber.New()
	v1 := app.Group("/api/v1")
	delivery.RegisterRBACRoutes(v1, uc, jwtManager)

	return app, jwtManager, token
}

func TestRBACHandler_GetRolesLookup(t *testing.T) {
	mockUC := &mockRBACUsecase{
		getRolesLookupFn: func(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*usecase.RoleLookupResponse, int64, error) {
			return []*usecase.RoleLookupResponse{
				{ID: "r-1", Code: "OWNER", Name: "Owner", Tier: "ORGANIZATION"},
				{ID: "r-2", Code: "CASHIER", Name: "Kasir POS", Tier: "OUTLET"},
			}, 2, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roles/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}

func TestRBACHandler_GetPermissionsLookup(t *testing.T) {
	mockUC := &mockRBACUsecase{
		getPermissionsLookupFn: func(ctx context.Context) ([]*usecase.PermissionGroupResponse, error) {
			return []*usecase.PermissionGroupResponse{
				{
					Group:     "pos",
					GroupName: "Point of Sale (POS)",
					Permissions: []*usecase.PermissionResponse{
						{ID: "p-1", Code: "pos:order:create", Group: "pos", Name: "Buat Pesanan"},
					},
				},
			}, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/permissions/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}

func TestRBACHandler_CreateRole(t *testing.T) {
	t.Run("success create", func(t *testing.T) {
		mockUC := &mockRBACUsecase{
			createRoleFn: func(ctx context.Context, input usecase.CreateRoleInput) (*usecase.RoleResponse, error) {
				return &usecase.RoleResponse{
					ID:   "new-role-id",
					Code: input.Code,
					Name: input.Name,
					Tier: input.Tier,
				}, nil
			},
		}

		app, _, token := setupTestApp(t, mockUC)

		payload := map[string]interface{}{
			"code":        "SUPERVISOR",
			"name":        "Store Supervisor",
			"description": "Supervisor Harian Cabang",
			"tier":        "OUTLET",
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("reject invalid body validation", func(t *testing.T) {
		mockUC := &mockRBACUsecase{}
		app, _, token := setupTestApp(t, mockUC)

		payload := map[string]interface{}{
			"code": "", // missing required
			"name": "",
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})
}
