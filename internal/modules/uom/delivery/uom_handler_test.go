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

	"github.com/hanjo-erp/hanjo/internal/modules/uom/delivery"
	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/uom/usecase"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockUoMUsecase struct {
	createUoMFn       func(ctx context.Context, orgID string, req usecase.CreateUoMRequest) (*usecase.UoMResponse, error)
	getUoMByIDFn      func(ctx context.Context, orgID string, id string) (*usecase.UoMResponse, error)
	listUoMsFn        func(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*usecase.UoMResponse, int64, error)
	listUoMLookupFn   func(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*usecase.UoMLookupResponse, error)
	updateUoMFn       func(ctx context.Context, orgID string, id string, req usecase.UpdateUoMRequest) (*usecase.UoMResponse, error)
	deleteUoMFn       func(ctx context.Context, orgID string, id string) error
	createConversionFn func(ctx context.Context, orgID string, req usecase.CreateConversionRequest) (*usecase.UoMConversionResponse, error)
	listConversionsFn func(ctx context.Context, orgID string) ([]*usecase.UoMConversionResponse, error)
	deleteConversionFn func(ctx context.Context, orgID string, id string) error
	convertQuantityFn func(ctx context.Context, orgID string, req usecase.ConvertQuantityRequest) (*usecase.ConvertQuantityResponse, error)
}

func (m *mockUoMUsecase) CreateUoM(ctx context.Context, orgID string, req usecase.CreateUoMRequest) (*usecase.UoMResponse, error) {
	if m.createUoMFn != nil {
		return m.createUoMFn(ctx, orgID, req)
	}
	return nil, nil
}

func (m *mockUoMUsecase) GetUoMByID(ctx context.Context, orgID string, id string) (*usecase.UoMResponse, error) {
	if m.getUoMByIDFn != nil {
		return m.getUoMByIDFn(ctx, orgID, id)
	}
	return nil, domain.ErrUoMNotFound
}

func (m *mockUoMUsecase) ListUoMs(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*usecase.UoMResponse, int64, error) {
	if m.listUoMsFn != nil {
		return m.listUoMsFn(ctx, orgID, filter, pagination)
	}
	return []*usecase.UoMResponse{}, 0, nil
}

func (m *mockUoMUsecase) ListUoMLookup(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*usecase.UoMLookupResponse, error) {
	if m.listUoMLookupFn != nil {
		return m.listUoMLookupFn(ctx, orgID, category)
	}
	return []*usecase.UoMLookupResponse{}, nil
}

func (m *mockUoMUsecase) UpdateUoM(ctx context.Context, orgID string, id string, req usecase.UpdateUoMRequest) (*usecase.UoMResponse, error) {
	if m.updateUoMFn != nil {
		return m.updateUoMFn(ctx, orgID, id, req)
	}
	return nil, nil
}

func (m *mockUoMUsecase) DeleteUoM(ctx context.Context, orgID string, id string) error {
	if m.deleteUoMFn != nil {
		return m.deleteUoMFn(ctx, orgID, id)
	}
	return nil
}

func (m *mockUoMUsecase) CreateConversion(ctx context.Context, orgID string, req usecase.CreateConversionRequest) (*usecase.UoMConversionResponse, error) {
	if m.createConversionFn != nil {
		return m.createConversionFn(ctx, orgID, req)
	}
	return nil, nil
}

func (m *mockUoMUsecase) ListConversions(ctx context.Context, orgID string) ([]*usecase.UoMConversionResponse, error) {
	if m.listConversionsFn != nil {
		return m.listConversionsFn(ctx, orgID)
	}
	return []*usecase.UoMConversionResponse{}, nil
}

func (m *mockUoMUsecase) DeleteConversion(ctx context.Context, orgID string, id string) error {
	if m.deleteConversionFn != nil {
		return m.deleteConversionFn(ctx, orgID, id)
	}
	return nil
}

func (m *mockUoMUsecase) ConvertQuantity(ctx context.Context, orgID string, req usecase.ConvertQuantityRequest) (*usecase.ConvertQuantityResponse, error) {
	if m.convertQuantityFn != nil {
		return m.convertQuantityFn(ctx, orgID, req)
	}
	return nil, nil
}

func setupTestApp(t *testing.T, uc usecase.UoMUsecase) (*fiber.App, *jwt.Manager, string) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtMgr := jwt.NewRSAManager(privKey, pubKey, "hanjo", "hanjo")

	app := fiber.New()
	v1 := app.Group("/api/v1")
	delivery.RegisterUoMRoutes(v1, uc, jwtMgr)

	token, err := jwtMgr.GenerateToken(jwt.TokenPayload{
		UserID:         "user-1",
		Role:           "OWNER",
		OrganizationID: "test-org-123",
		Duration:       time.Hour,
	})
	require.NoError(t, err)

	return app, jwtMgr, token
}

func TestUoMHandler_GetLookup_Success(t *testing.T) {
	mockUC := &mockUoMUsecase{
		listUoMLookupFn: func(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*usecase.UoMLookupResponse, error) {
			return []*usecase.UoMLookupResponse{
				{ID: "uom-1", Name: "Gram", Symbol: "g", Category: "WEIGHT", DecimalPlaces: 2},
				{ID: "uom-2", Name: "Milliliter", Symbol: "ml", Category: "VOLUME", DecimalPlaces: 2},
			}, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/uoms/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}

func TestUoMHandler_Create_ValidationFailure(t *testing.T) {
	mockUC := &mockUoMUsecase{}
	app, _, token := setupTestApp(t, mockUC)

	// Missing required fields
	payload := map[string]string{
		"name": "",
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uoms", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)
}

func TestUoMHandler_ConvertQuantity_Success(t *testing.T) {
	mockUC := &mockUoMUsecase{
		convertQuantityFn: func(ctx context.Context, orgID string, req usecase.ConvertQuantityRequest) (*usecase.ConvertQuantityResponse, error) {
			return &usecase.ConvertQuantityResponse{
				FromUoM: domain.UoMSummary{
					ID:       "uom-kg",
					Name:     "Kilogram",
					Symbol:   "kg",
					Category: "WEIGHT",
				},
				ToUoM: domain.UoMSummary{
					ID:       "uom-g",
					Name:     "Gram",
					Symbol:   "g",
					Category: "WEIGHT",
				},
				OriginalQuantity:  2.0,
				ConvertedQuantity: 2000.0,
				Factor:            1000.0,
			}, nil
		},
	}

	app, _, token := setupTestApp(t, mockUC)

	payload := map[string]interface{}{
		"from_uom_id": "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d",
		"to_uom_id":   "c4d1e2f3-a4b5-4c6d-8e9f-0a1b2c3d4e5f",
		"quantity":    2.0,
	}
	b, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/uom-conversions/convert", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
}
