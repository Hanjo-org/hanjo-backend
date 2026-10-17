package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	authdomain "github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/auth/usecase"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockUserRepository struct {
	getByIDFn                func(ctx context.Context, id string) (*userdomain.User, error)
	getByEmailFn             func(ctx context.Context, email string) (*userdomain.User, error)
	getByUsernameFn          func(ctx context.Context, orgID, username string) (*userdomain.User, error)
	getByOutletAndUsernameFn func(ctx context.Context, outletID, username string) (*userdomain.User, error)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepository) GetByIDs(ctx context.Context, ids []string) ([]*userdomain.User, error) {
	return nil, nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, orgID, username string) (*userdomain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, orgID, username)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepository) GetByOutletAndUsername(ctx context.Context, outletID, username string) (*userdomain.User, error) {
	if m.getByOutletAndUsernameFn != nil {
		return m.getByOutletAndUsernameFn(ctx, outletID, username)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepository) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return nil, 0, nil
}

func (m *mockUserRepository) Create(ctx context.Context, user *userdomain.User) error {
	return nil
}

func (m *mockUserRepository) Update(ctx context.Context, user *userdomain.User) error {
	return nil
}

func (m *mockUserRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func setupTestUsecase(t *testing.T, repo userdomain.UserRepository) (usecase.AuthUsecase, *jwt.Manager) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	require.NoError(t, err)

	jwtManager := jwt.NewRSAManager(privKey, pubKey, "hanjo-test", "hanjo-test")
	uc := usecase.NewAuthUsecase(
		repo,
		jwtManager,
		15*time.Minute,
		7*24*time.Hour,
		zap.NewNop(),
	)
	return uc, jwtManager
}

func TestAuthUsecase_Login(t *testing.T) {
	hashedPwd, err := password.Hash("Secret123!")
	require.NoError(t, err)

	activeUser := &userdomain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Username:       "budi",
		Email:          "budi@example.com",
		Password:       hashedPwd,
		Status:         "ACTIVE",
	}

	t.Run("Success", func(t *testing.T) {
		repo := &mockUserRepository{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return activeUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.Login(context.Background(), usecase.LoginInput{
			Identifier: "budi@example.com",
			Password:   "Secret123!",
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "user-1", out.User.ID)
		assert.NotEmpty(t, out.Tokens.AccessToken)
		assert.NotEmpty(t, out.Tokens.RefreshToken)
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo := &mockUserRepository{}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.Login(context.Background(), usecase.LoginInput{
			Identifier: "unknown@example.com",
			Password:   "Secret123!",
		})
		assert.Nil(t, out)
		assert.ErrorIs(t, err, authdomain.ErrInvalidCredentials)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		repo := &mockUserRepository{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return activeUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.Login(context.Background(), usecase.LoginInput{
			Identifier: "budi@example.com",
			Password:   "WrongPassword!",
		})
		assert.Nil(t, out)
		assert.ErrorIs(t, err, authdomain.ErrInvalidCredentials)
	})

	t.Run("Inactive User", func(t *testing.T) {
		inactiveUser := *activeUser
		inactiveUser.Status = "INACTIVE"

		repo := &mockUserRepository{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return &inactiveUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.Login(context.Background(), usecase.LoginInput{
			Identifier: "budi@example.com",
			Password:   "Secret123!",
		})
		assert.Nil(t, out)
		assert.ErrorIs(t, err, authdomain.ErrAccountInactive)
	})
}

func TestAuthUsecase_POSLogin(t *testing.T) {
	hashedPIN, err := password.Hash("123456")
	require.NoError(t, err)

	cashierUser := &userdomain.User{
		ID:             "cashier-1",
		OrganizationID: "org-1",
		OutletID:       ptr.Ptr("outlet-1"),
		Username:       "kasir1",
		Email:          "kasir1@example.com",
		Pin:            &hashedPIN,
		Status:         "ACTIVE",
	}

	t.Run("Success", func(t *testing.T) {
		repo := &mockUserRepository{
			getByOutletAndUsernameFn: func(ctx context.Context, outletID, username string) (*userdomain.User, error) {
				return cashierUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.POSLogin(context.Background(), usecase.POSLoginInput{
			OutletID: "outlet-1",
			Username: "kasir1",
			PIN:      "123456",
		})
		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.Equal(t, "cashier-1", out.User.ID)
		assert.NotEmpty(t, out.Tokens.AccessToken)
	})

	t.Run("Invalid PIN", func(t *testing.T) {
		repo := &mockUserRepository{
			getByOutletAndUsernameFn: func(ctx context.Context, outletID, username string) (*userdomain.User, error) {
				return cashierUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		out, err := uc.POSLogin(context.Background(), usecase.POSLoginInput{
			OutletID: "outlet-1",
			Username: "kasir1",
			PIN:      "000000",
		})
		assert.Nil(t, out)
		assert.ErrorIs(t, err, authdomain.ErrInvalidPIN)
	})
}

func TestAuthUsecase_RefreshToken(t *testing.T) {
	activeUser := &userdomain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Username:       "budi",
		Email:          "budi@example.com",
		Status:         "ACTIVE",
	}

	repo := &mockUserRepository{
		getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
			if id == "user-1" {
				return activeUser, nil
			}
			return nil, userdomain.ErrUserNotFound
		},
	}
	uc, jwtManager := setupTestUsecase(t, repo)

	validRefreshToken, err := jwtManager.GenerateToken(jwt.TokenPayload{
		UserID:         "user-1",
		Email:          "budi@example.com",
		OrganizationID: "org-1",
		Duration:       7 * 24 * time.Hour,
	})
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		tokens, err := uc.RefreshToken(context.Background(), validRefreshToken)
		assert.NoError(t, err)
		assert.NotNil(t, tokens)
		assert.NotEmpty(t, tokens.AccessToken)
		assert.NotEmpty(t, tokens.RefreshToken)
	})

	t.Run("Invalid Token String", func(t *testing.T) {
		tokens, err := uc.RefreshToken(context.Background(), "invalid.token.string")
		assert.Nil(t, tokens)
		assert.Error(t, err)
	})
}

func TestAuthUsecase_Me(t *testing.T) {
	activeUser := &userdomain.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Username:       "budi",
		Email:          "budi@example.com",
		Status:         "ACTIVE",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		repo := &mockUserRepository{
			getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
				return activeUser, nil
			},
		}
		uc, _ := setupTestUsecase(t, repo)

		profile, err := uc.Me(context.Background(), "user-1")
		assert.NoError(t, err)
		assert.NotNil(t, profile)
		assert.Equal(t, "user-1", profile.ID)
		assert.Equal(t, "budi", profile.Username)
	})

	t.Run("User Not Found", func(t *testing.T) {
		repo := &mockUserRepository{}
		uc, _ := setupTestUsecase(t, repo)

		profile, err := uc.Me(context.Background(), "unknown-user")
		assert.Nil(t, profile)
		assert.ErrorIs(t, err, userdomain.ErrUserNotFound)
	})
}
