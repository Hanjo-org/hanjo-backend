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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/delivery"
	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/supplier/usecase"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockSupplierUsecase struct {
	createSupplierFn      func(ctx context.Context, orgID string, req usecase.CreateSupplierRequest) (*usecase.SupplierResponse, error)
	getSupplierByIDFn     func(ctx context.Context, orgID, id string) (*usecase.SupplierResponse, error)
	listSuppliersFn       func(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]usecase.SupplierResponse, int64, error)
	listLookupSuppliersFn func(ctx context.Context, orgID string) ([]usecase.SupplierLookupResponse, error)
	updateSupplierFn      func(ctx context.Context, orgID, id string, req usecase.UpdateSupplierRequest) (*usecase.SupplierResponse, error)
	deleteSupplierFn      func(ctx context.Context, orgID, id string) error
}

func (m *mockSupplierUsecase) CreateSupplier(ctx context.Context, orgID string, req usecase.CreateSupplierRequest) (*usecase.SupplierResponse, error) {
	if m.createSupplierFn != nil {
		return m.createSupplierFn(ctx, orgID, req)
	}
	return nil, nil
}

func (m *mockSupplierUsecase) GetSupplierByID(ctx context.Context, orgID, id string) (*usecase.SupplierResponse, error) {
	if m.getSupplierByIDFn != nil {
		return m.getSupplierByIDFn(ctx, orgID, id)
	}
	return nil, domain.ErrSupplierNotFound
}

func (m *mockSupplierUsecase) ListSuppliers(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]usecase.SupplierResponse, int64, error) {
	if m.listSuppliersFn != nil {
		return m.listSuppliersFn(ctx, orgID, filter, pagination)
	}
	return []usecase.SupplierResponse{}, 0, nil
}

func (m *mockSupplierUsecase) ListLookupSuppliers(ctx context.Context, orgID string) ([]usecase.SupplierLookupResponse, error) {
	if m.listLookupSuppliersFn != nil {
		return m.listLookupSuppliersFn(ctx, orgID)
	}
	return []usecase.SupplierLookupResponse{}, nil
}

func (m *mockSupplierUsecase) UpdateSupplier(ctx context.Context, orgID, id string, req usecase.UpdateSupplierRequest) (*usecase.SupplierResponse, error) {
	if m.updateSupplierFn != nil {
		return m.updateSupplierFn(ctx, orgID, id, req)
	}
	return nil, nil
}

func (m *mockSupplierUsecase) DeleteSupplier(ctx context.Context, orgID, id string) error {
	if m.deleteSupplierFn != nil {
		return m.deleteSupplierFn(ctx, orgID, id)
	}
	return nil
}

func setupTestApp(t *testing.T, uc usecase.SupplierUsecase) (*fiber.App, *jwt.Manager, string) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtMgr := jwt.NewRSAManager(privKey, pubKey, "hanjo", "hanjo")

	app := fiber.New()
	v1 := app.Group("/api/v1")
	delivery.RegisterSupplierRoutes(v1, uc, jwtMgr)

	token, err := jwtMgr.GenerateToken(jwt.TokenPayload{
		UserID:         "user-1",
		Role:           "OWNER",
		OrganizationID: "test-org-123",
		Duration:       time.Hour,
	})
	require.NoError(t, err)

	return app, jwtMgr, token
}

func TestSupplierHandler_GetLookup_Success(t *testing.T) {
	mockUC := &mockSupplierUsecase{
		listLookupSuppliersFn: func(ctx context.Context, orgID string) ([]usecase.SupplierLookupResponse, error) {
			return []usecase.SupplierLookupResponse{
				{ID: "supp-1", Name: "Diamond Milk", Code: ptr.Ptr("SUPP-01")},
				{ID: "supp-2", Name: "Toffin", Code: ptr.Ptr("SUPP-02")},
			}, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/suppliers/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}

func TestSupplierHandler_List_Success(t *testing.T) {
	mockUC := &mockSupplierUsecase{
		listSuppliersFn: func(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]usecase.SupplierResponse, int64, error) {
			return []usecase.SupplierResponse{
				{ID: "supp-1", OrganizationID: orgID, Name: "Diamond Milk", Status: "ACTIVE"},
			}, 1, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/suppliers?page=1&per_page=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestSupplierHandler_Create_ValidationFailure(t *testing.T) {
	mockUC := &mockSupplierUsecase{}
	app, _, token := setupTestApp(t, mockUC)

	// Missing required field 'name'
	payload := map[string]string{
		"code": "SUPP-01",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/suppliers", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestSupplierHandler_Create_Success(t *testing.T) {
	mockUC := &mockSupplierUsecase{
		createSupplierFn: func(ctx context.Context, orgID string, req usecase.CreateSupplierRequest) (*usecase.SupplierResponse, error) {
			return &usecase.SupplierResponse{
				ID:             "new-supp-id",
				OrganizationID: orgID,
				Name:           req.Name,
				Status:         "ACTIVE",
			}, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	payload := map[string]string{
		"name": "PT Diamond Cold Storage",
		"code": "SUPP-01",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/suppliers", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestSupplierHandler_GetByID_NotFound(t *testing.T) {
	mockUC := &mockSupplierUsecase{
		getSupplierByIDFn: func(ctx context.Context, orgID, id string) (*usecase.SupplierResponse, error) {
			return nil, domain.ErrSupplierNotFound
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/suppliers/non-existent-id", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestSupplierHandler_Delete_Success(t *testing.T) {
	mockUC := &mockSupplierUsecase{
		deleteSupplierFn: func(ctx context.Context, orgID, id string) error {
			return nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/suppliers/supp-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}
