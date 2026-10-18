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

	"github.com/hanjo-erp/hanjo/internal/modules/user/delivery"
	"github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/user/usecase"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockUserUsecase struct {
	listFn         func(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserWithRelations, int64, error)
	getLookupFn    func(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error)
	getByIDFn      func(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error)
	createFn       func(ctx context.Context, input usecase.CreateUserInput) (*domain.UserWithRelations, error)
	updateFn       func(ctx context.Context, orgID, id string, input usecase.UpdateUserInput) (*domain.UserWithRelations, error)
	updateStatusFn func(ctx context.Context, orgID, id string, status string) (*domain.UserWithRelations, error)
	deleteFn       func(ctx context.Context, orgID, id string) error
}

func (m *mockUserUsecase) List(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserWithRelations, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, params)
	}
	return nil, 0, nil
}

func (m *mockUserUsecase) GetLookup(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
	if m.getLookupFn != nil {
		return m.getLookupFn(ctx, filter, params)
	}
	return nil, 0, nil
}

func (m *mockUserUsecase) GetByID(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, orgID, id)
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserUsecase) Create(ctx context.Context, input usecase.CreateUserInput) (*domain.UserWithRelations, error) {
	if m.createFn != nil {
		return m.createFn(ctx, input)
	}
	return nil, nil
}

func (m *mockUserUsecase) Update(ctx context.Context, orgID, id string, input usecase.UpdateUserInput) (*domain.UserWithRelations, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, orgID, id, input)
	}
	return nil, nil
}

func (m *mockUserUsecase) UpdateStatus(ctx context.Context, orgID, id string, status string) (*domain.UserWithRelations, error) {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, orgID, id, status)
	}
	return nil, nil
}

func (m *mockUserUsecase) Delete(ctx context.Context, orgID, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, orgID, id)
	}
	return nil
}

func setupTestApp(t *testing.T, uc usecase.UserUsecase) (*fiber.App, *jwt.Manager, string) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtManager := jwt.NewRSAManager(privKey, pubKey, "hanjo-test", "hanjo-test")

	app := fiber.New()
	v1 := app.Group("/api/v1")
	delivery.RegisterUserRoutes(v1, uc, jwtManager)

	token, err := jwtManager.GenerateToken(jwt.TokenPayload{
		UserID:         uuid.NewString(),
		Email:          "owner@harmoni.id",
		Role:           "OWNER",
		OrganizationID: "org-123",
		Duration:       time.Hour,
	})
	require.NoError(t, err)

	return app, jwtManager, token
}

func TestUserHandler_List(t *testing.T) {
	orgID := "org-123"
	uc := &mockUserUsecase{
		listFn: func(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserWithRelations, int64, error) {
			assert.Equal(t, orgID, filter.OrganizationID)
			return []*domain.UserWithRelations{
				{
					User: &domain.User{
						ID:             "u-1",
						OrganizationID: orgID,
						Username:       "kasir1",
						Email:          "kasir1@harmoni.id",
						Status:         "ACTIVE",
					},
					Organization: &domain.OrganizationSummary{ID: orgID, Name: "PT Harmoni"},
					Brand:        &domain.BrandSummary{ID: "b-1", Name: "Daily Coffee"},
					Outlet:       &domain.OutletSummary{ID: "o-1", Name: "Kemang"},
				},
			}, 1, nil
		},
	}

	app, _, token := setupTestApp(t, uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&per_page=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	assert.NotNil(t, body["meta"])
	metaMap := body["meta"].(map[string]any)
	assert.NotNil(t, metaMap["pagination"])
}

func TestUserHandler_GetLookup(t *testing.T) {
	orgID := "org-123"
	uc := &mockUserUsecase{
		getLookupFn: func(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
			assert.Equal(t, orgID, filter.OrganizationID)
			return []*domain.UserLookup{
				{ID: "u-1", Name: "kasir.kemang"},
				{ID: "u-2", Name: "barista.kemang"},
			}, 2, nil
		},
	}

	app, _, token := setupTestApp(t, uc)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/lookup", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.True(t, body["success"].(bool))
	dataList := body["data"].([]any)
	assert.Len(t, dataList, 2)
	assert.NotNil(t, body["meta"])
	metaMap := body["meta"].(map[string]any)
	assert.NotNil(t, metaMap["pagination"])
}

func TestUserHandler_GetByID(t *testing.T) {
	userID := "u-1"
	expectedOrgID := "org-123"

	t.Run("found user", func(t *testing.T) {
		uc := &mockUserUsecase{
			getByIDFn: func(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error) {
				assert.Equal(t, expectedOrgID, orgID)
				return &domain.UserWithRelations{
					User: &domain.User{
						ID:             userID,
						OrganizationID: orgID,
						Username:       "kasir1",
						Email:          "kasir1@harmoni.id",
					},
					Organization: &domain.OrganizationSummary{ID: orgID, Name: "PT Harmoni"},
				}, nil
			},
		}

		app, _, token := setupTestApp(t, uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/"+userID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("user not found", func(t *testing.T) {
		uc := &mockUserUsecase{
			getByIDFn: func(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error) {
				return nil, domain.ErrUserNotFound
			},
		}

		app, _, token := setupTestApp(t, uc)

		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/non-existent", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

func TestUserHandler_Create(t *testing.T) {
	t.Run("success create", func(t *testing.T) {
		uc := &mockUserUsecase{
			createFn: func(ctx context.Context, input usecase.CreateUserInput) (*domain.UserWithRelations, error) {
				return &domain.UserWithRelations{
					User: &domain.User{
						ID:             "u-new",
						OrganizationID: input.OrganizationID,
						Username:       input.Username,
						Email:          input.Email,
						Status:         "ACTIVE",
					},
				}, nil
			},
		}

		app, _, token := setupTestApp(t, uc)

		payload := delivery.CreateUserRequest{
			Username: "barista1",
			Email:    "barista1@harmoni.id",
			Password: "Password123!",
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	})

	t.Run("validation failure invalid email", func(t *testing.T) {
		uc := &mockUserUsecase{}
		app, _, token := setupTestApp(t, uc)

		payload := map[string]any{
			"username": "barista1",
			"email":    "not-an-email",
			"password": "Password123!",
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("conflict duplicate email", func(t *testing.T) {
		uc := &mockUserUsecase{
			createFn: func(ctx context.Context, input usecase.CreateUserInput) (*domain.UserWithRelations, error) {
				return nil, domain.ErrEmailAlreadyExists
			},
		}

		app, _, token := setupTestApp(t, uc)

		payload := delivery.CreateUserRequest{
			Username: "barista1",
			Email:    "existing@harmoni.id",
			Password: "Password123!",
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)
	})
}

func TestUserHandler_Update(t *testing.T) {
	uc := &mockUserUsecase{
		updateFn: func(ctx context.Context, orgID, id string, input usecase.UpdateUserInput) (*domain.UserWithRelations, error) {
			return &domain.UserWithRelations{
				User: &domain.User{
					ID:             id,
					OrganizationID: orgID,
					Username:       "updated_username",
				},
			}, nil
		},
	}

	app, _, token := setupTestApp(t, uc)

	payload := map[string]any{
		"username": "updated_username",
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/u-1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestUserHandler_UpdateStatus(t *testing.T) {
	t.Run("success update status", func(t *testing.T) {
		uc := &mockUserUsecase{
			updateStatusFn: func(ctx context.Context, orgID, id string, status string) (*domain.UserWithRelations, error) {
				return &domain.UserWithRelations{
					User: &domain.User{
						ID:             id,
						OrganizationID: orgID,
						Status:         status,
					},
				}, nil
			},
		}

		app, _, token := setupTestApp(t, uc)

		payload := delivery.UpdateUserStatusRequest{Status: "INACTIVE"}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/u-1/status", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("invalid status", func(t *testing.T) {
		uc := &mockUserUsecase{}
		app, _, token := setupTestApp(t, uc)

		payload := map[string]any{"status": "UNKNOWN_STATUS"}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/u-1/status", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})
}

func TestUserHandler_Delete(t *testing.T) {
	uc := &mockUserUsecase{
		deleteFn: func(ctx context.Context, orgID, id string) error {
			return nil
		},
	}

	app, _, token := setupTestApp(t, uc)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/u-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
