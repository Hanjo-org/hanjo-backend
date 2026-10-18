package delivery_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanjo-erp/hanjo/internal/modules/auth/delivery"
	authdomain "github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/auth/usecase"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
)

type mockAuthUsecase struct {
	loginFn          func(ctx context.Context, input usecase.LoginInput) (*usecase.AuthOutput, error)
	posLoginFn       func(ctx context.Context, input usecase.POSLoginInput) (*usecase.AuthOutput, error)
	refreshTokenFn   func(ctx context.Context, refreshTokenString string) (*authdomain.TokenPair, error)
	meFn             func(ctx context.Context, userID string) (*usecase.UserProfileOutput, error)
	changePasswordFn func(ctx context.Context, input usecase.ChangePasswordInput) error
	changePINFn      func(ctx context.Context, input usecase.ChangePINInput) error
}

func (m *mockAuthUsecase) Login(ctx context.Context, input usecase.LoginInput) (*usecase.AuthOutput, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, input)
	}
	return nil, authdomain.ErrInvalidCredentials
}

func (m *mockAuthUsecase) POSLogin(ctx context.Context, input usecase.POSLoginInput) (*usecase.AuthOutput, error) {
	if m.posLoginFn != nil {
		return m.posLoginFn(ctx, input)
	}
	return nil, authdomain.ErrInvalidPIN
}

func (m *mockAuthUsecase) RefreshToken(ctx context.Context, refreshTokenString string) (*authdomain.TokenPair, error) {
	if m.refreshTokenFn != nil {
		return m.refreshTokenFn(ctx, refreshTokenString)
	}
	return nil, authdomain.ErrInvalidToken
}

func (m *mockAuthUsecase) Me(ctx context.Context, userID string) (*usecase.UserProfileOutput, error) {
	if m.meFn != nil {
		return m.meFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockAuthUsecase) ChangePassword(ctx context.Context, input usecase.ChangePasswordInput) error {
	if m.changePasswordFn != nil {
		return m.changePasswordFn(ctx, input)
	}
	return nil
}

func (m *mockAuthUsecase) ChangePIN(ctx context.Context, input usecase.ChangePINInput) error {
	if m.changePINFn != nil {
		return m.changePINFn(ctx, input)
	}
	return nil
}

func setupTestApp(t *testing.T, uc usecase.AuthUsecase) (*fiber.App, *jwt.Manager) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtManager := jwt.NewRSAManager(privKey, pubKey, "hanjo-test", "hanjo-test")

	app := fiber.New()
	v1 := app.Group("/api/v1")
	delivery.RegisterAuthRoutes(v1, uc, jwtManager, "refresh_token", false)

	return app, jwtManager
}

func TestAuthHandler_Login(t *testing.T) {
	uc := &mockAuthUsecase{
		loginFn: func(ctx context.Context, input usecase.LoginInput) (*usecase.AuthOutput, error) {
			if input.Identifier == "admin@example.com" && input.Password == "Secret123!" {
				return &usecase.AuthOutput{
					User: usecase.UserOutput{
						ID:             "u-1",
						OrganizationID: "org-1",
						Username:       "admin",
						Email:          "admin@example.com",
						Status:         "ACTIVE",
					},
					Tokens: authdomain.TokenPair{
						AccessToken:  "mock-access-token",
						RefreshToken: "mock-refresh-token",
						TokenType:    "Bearer",
						ExpiresIn:    900,
						ExpiresAt:    time.Now().Add(15 * time.Minute),
					},
				}, nil
			}
			return nil, authdomain.ErrInvalidCredentials
		},
	}

	app, _ := setupTestApp(t, uc)

	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": "admin@example.com",
			"password":   "Secret123!",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Validation Failure: Missing Password", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": "admin@example.com",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Invalid Credentials", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"identifier": "admin@example.com",
			"password":   "WrongPass123!",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Too Many Attempts: 429 Status", func(t *testing.T) {
		lockedUc := &mockAuthUsecase{
			loginFn: func(ctx context.Context, input usecase.LoginInput) (*usecase.AuthOutput, error) {
				return nil, authdomain.ErrTooManyAttempts
			},
		}
		lockedApp, _ := setupTestApp(t, lockedUc)

		body, _ := json.Marshal(map[string]string{
			"identifier": "locked@example.com",
			"password":   "AnyPassword123!",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := lockedApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})
}

func TestAuthHandler_POSLogin(t *testing.T) {
	uc := &mockAuthUsecase{
		posLoginFn: func(ctx context.Context, input usecase.POSLoginInput) (*usecase.AuthOutput, error) {
			if input.PIN == "123456" {
				return &usecase.AuthOutput{
					User: usecase.UserOutput{
						ID:             "c-1",
						OrganizationID: "org-1",
						OutletID:       &input.OutletID,
						Username:       input.Username,
						Status:         "ACTIVE",
					},
					Tokens: authdomain.TokenPair{
						AccessToken: "mock-pos-token",
					},
				}, nil
			}
			return nil, authdomain.ErrInvalidPIN
		},
	}

	app, _ := setupTestApp(t, uc)

	t.Run("Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"outlet_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			"username":  "kasir1",
			"pin":       "123456",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/pos/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Validation Failure: Non-UUID Outlet ID", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"outlet_id": "invalid-uuid",
			"username":  "kasir1",
			"pin":       "123456",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/pos/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("Too Many Attempts: 429 Status", func(t *testing.T) {
		lockedUc := &mockAuthUsecase{
			posLoginFn: func(ctx context.Context, input usecase.POSLoginInput) (*usecase.AuthOutput, error) {
				return nil, authdomain.ErrTooManyAttempts
			},
		}
		lockedApp, _ := setupTestApp(t, lockedUc)

		body, _ := json.Marshal(map[string]string{
			"outlet_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
			"username":  "kasir_locked",
			"pin":       "123456",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/pos/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := lockedApp.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	})
}

func TestAuthHandler_ProtectedRoutes(t *testing.T) {
	uc := &mockAuthUsecase{
		meFn: func(ctx context.Context, userID string) (*usecase.UserProfileOutput, error) {
			return &usecase.UserProfileOutput{
				ID:             userID,
				OrganizationID: "org-1",
				Username:       "budi",
				Email:          "budi@example.com",
				Status:         "ACTIVE",
			}, nil
		},
	}

	app, jwtManager := setupTestApp(t, uc)

	token, err := jwtManager.GenerateToken(jwt.TokenPayload{
		UserID:         "user-1",
		Email:          "budi@example.com",
		OrganizationID: "org-1",
		Duration:       15 * time.Minute,
	})
	require.NoError(t, err)

	t.Run("Me: Unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Me: Authorized with valid Bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Logout: Authorized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ChangePassword: Success", func(t *testing.T) {
		uc.changePasswordFn = func(ctx context.Context, input usecase.ChangePasswordInput) error {
			assert.Equal(t, "user-1", input.UserID)
			assert.Equal(t, "CurrentSecret123", input.CurrentPassword)
			assert.Equal(t, "NewSecret456", input.NewPassword)
			return nil
		}

		body := `{"current_password":"CurrentSecret123","new_password":"NewSecret456"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/password", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ChangePassword: Current Password Mismatch", func(t *testing.T) {
		uc.changePasswordFn = func(ctx context.Context, input usecase.ChangePasswordInput) error {
			return authdomain.ErrCurrentPasswordMismatch
		}

		body := `{"current_password":"WrongSecret123","new_password":"NewSecret456"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/password", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("ChangePassword: Validation Error when new password too short", func(t *testing.T) {
		body := `{"current_password":"CurrentSecret123","new_password":"short"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/password", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})

	t.Run("ChangePIN: Success", func(t *testing.T) {
		uc.changePINFn = func(ctx context.Context, input usecase.ChangePINInput) error {
			assert.Equal(t, "user-1", input.UserID)
			assert.Equal(t, "123456", *input.CurrentPIN)
			assert.Equal(t, "654321", input.NewPIN)
			return nil
		}

		body := `{"current_pin":"123456","new_pin":"654321"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/pin", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ChangePIN: Current PIN Mismatch", func(t *testing.T) {
		uc.changePINFn = func(ctx context.Context, input usecase.ChangePINInput) error {
			return authdomain.ErrCurrentPINMismatch
		}

		body := `{"current_pin":"000000","new_pin":"654321"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/pin", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("ChangePIN: Validation Error when PIN non-numeric", func(t *testing.T) {
		body := `{"new_pin":"abcdef"}`
		req := httptest.NewRequest(http.MethodPut, "/api/v1/auth/me/pin", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	})
}
