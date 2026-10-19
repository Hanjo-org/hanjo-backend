package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	orgdomain "github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	rbacdomain "github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/user/usecase"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Mock Repositories
type mockUserRepo struct {
	getByIDFn        func(ctx context.Context, id string) (*userdomain.User, error)
	getByIDsFn       func(ctx context.Context, ids []string) ([]*userdomain.User, error)
	getByEmailFn     func(ctx context.Context, email string) (*userdomain.User, error)
	getByUsernameFn  func(ctx context.Context, orgID, username string) (*userdomain.User, error)
	getByOutletAndFn func(ctx context.Context, outletID, username string) (*userdomain.User, error)
	listFn           func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error)
	getLookupFn      func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.UserLookup, int64, error)
	createFn         func(ctx context.Context, user *userdomain.User) error
	updateFn         func(ctx context.Context, user *userdomain.User) error
	deleteFn         func(ctx context.Context, id string) error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id string) (*userdomain.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepo) GetByIDs(ctx context.Context, ids []string) ([]*userdomain.User, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*userdomain.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, orgID, username string) (*userdomain.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, orgID, username)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepo) GetByOutletAndUsername(ctx context.Context, outletID, username string) (*userdomain.User, error) {
	if m.getByOutletAndFn != nil {
		return m.getByOutletAndFn(ctx, outletID, username)
	}
	return nil, userdomain.ErrUserNotFound
}

func (m *mockUserRepo) List(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, params)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return m.List(ctx, userdomain.UserFilter{OrganizationID: orgID}, params)
}

func (m *mockUserRepo) ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*userdomain.User, int64, error) {
	return m.List(ctx, userdomain.UserFilter{OutletID: &outletID}, params)
}

func (m *mockUserRepo) GetLookup(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.UserLookup, int64, error) {
	if m.getLookupFn != nil {
		return m.getLookupFn(ctx, filter, params)
	}
	return nil, 0, nil
}

func (m *mockUserRepo) Create(ctx context.Context, user *userdomain.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Update(ctx context.Context, user *userdomain.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockOrgRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*orgdomain.Organization, error)
	getByIDsFn  func(ctx context.Context, ids []string) ([]*orgdomain.Organization, error)
	getBySlugFn func(ctx context.Context, slug string) (*orgdomain.Organization, error)
	listFn      func(ctx context.Context, params response.PaginationParams) ([]*orgdomain.Organization, int64, error)
	createFn    func(ctx context.Context, org *orgdomain.Organization) error
	updateFn    func(ctx context.Context, org *orgdomain.Organization) error
	deleteFn    func(ctx context.Context, id string) error
}

func (m *mockOrgRepo) GetByID(ctx context.Context, id string) (*orgdomain.Organization, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, orgdomain.ErrOrganizationNotFound
}

func (m *mockOrgRepo) GetByIDs(ctx context.Context, ids []string) ([]*orgdomain.Organization, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*orgdomain.Organization, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, slug)
	}
	return nil, orgdomain.ErrOrganizationNotFound
}

func (m *mockOrgRepo) List(ctx context.Context, params response.PaginationParams) ([]*orgdomain.Organization, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, params)
	}
	return nil, 0, nil
}

func (m *mockOrgRepo) Create(ctx context.Context, org *orgdomain.Organization) error {
	return nil
}

func (m *mockOrgRepo) Update(ctx context.Context, org *orgdomain.Organization) error {
	return nil
}

func (m *mockOrgRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockBrandRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*orgdomain.Brand, error)
	getByIDsFn  func(ctx context.Context, ids []string) ([]*orgdomain.Brand, error)
	getBySlugFn func(ctx context.Context, orgID, slug string) (*orgdomain.Brand, error)
	listByOrgFn func(ctx context.Context, orgID string, params response.PaginationParams) ([]*orgdomain.Brand, int64, error)
	createFn    func(ctx context.Context, brand *orgdomain.Brand) error
	updateFn    func(ctx context.Context, brand *orgdomain.Brand) error
	deleteFn    func(ctx context.Context, id string) error
}

func (m *mockBrandRepo) GetByID(ctx context.Context, id string) (*orgdomain.Brand, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, orgdomain.ErrBrandNotFound
}

func (m *mockBrandRepo) GetByIDs(ctx context.Context, ids []string) ([]*orgdomain.Brand, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockBrandRepo) GetBySlug(ctx context.Context, orgID, slug string) (*orgdomain.Brand, error) {
	if m.getBySlugFn != nil {
		return m.getBySlugFn(ctx, orgID, slug)
	}
	return nil, orgdomain.ErrBrandNotFound
}

func (m *mockBrandRepo) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*orgdomain.Brand, int64, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, params)
	}
	return nil, 0, nil
}

func (m *mockBrandRepo) Create(ctx context.Context, brand *orgdomain.Brand) error {
	return nil
}

func (m *mockBrandRepo) Update(ctx context.Context, brand *orgdomain.Brand) error {
	return nil
}

func (m *mockBrandRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockOutletRepo struct {
	getByIDFn   func(ctx context.Context, id string) (*orgdomain.Outlet, error)
	getByIDsFn  func(ctx context.Context, ids []string) ([]*orgdomain.Outlet, error)
	getByCodeFn func(ctx context.Context, orgID, code string) (*orgdomain.Outlet, error)
	listBrandFn func(ctx context.Context, brandID string, params response.PaginationParams) ([]*orgdomain.Outlet, int64, error)
	listOrgFn   func(ctx context.Context, orgID string, params response.PaginationParams) ([]*orgdomain.Outlet, int64, error)
	createFn    func(ctx context.Context, outlet *orgdomain.Outlet) error
	updateFn    func(ctx context.Context, outlet *orgdomain.Outlet) error
	deleteFn    func(ctx context.Context, id string) error
}

func (m *mockOutletRepo) GetByID(ctx context.Context, id string) (*orgdomain.Outlet, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, orgdomain.ErrOutletNotFound
}

func (m *mockOutletRepo) GetByIDs(ctx context.Context, ids []string) ([]*orgdomain.Outlet, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockOutletRepo) GetByCode(ctx context.Context, orgID, code string) (*orgdomain.Outlet, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, orgID, code)
	}
	return nil, orgdomain.ErrOutletNotFound
}

func (m *mockOutletRepo) ListByBrand(ctx context.Context, brandID string, params response.PaginationParams) ([]*orgdomain.Outlet, int64, error) {
	if m.listBrandFn != nil {
		return m.listBrandFn(ctx, brandID, params)
	}
	return nil, 0, nil
}

func (m *mockOutletRepo) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*orgdomain.Outlet, int64, error) {
	if m.listOrgFn != nil {
		return m.listOrgFn(ctx, orgID, params)
	}
	return nil, 0, nil
}

func (m *mockOutletRepo) Create(ctx context.Context, outlet *orgdomain.Outlet) error {
	return nil
}

func (m *mockOutletRepo) Update(ctx context.Context, outlet *orgdomain.Outlet) error {
	return nil
}

func (m *mockOutletRepo) Delete(ctx context.Context, id string) error {
	return nil
}

type mockRoleRepo struct {
	getByIDFn                      func(ctx context.Context, id string) (*rbacdomain.Role, error)
	getByIDsFn                     func(ctx context.Context, ids []string) ([]*rbacdomain.Role, error)
	getByCodeFn                    func(ctx context.Context, orgID *string, code string) (*rbacdomain.Role, error)
	listFn                         func(ctx context.Context, filter rbacdomain.RoleFilter, params response.PaginationParams) ([]*rbacdomain.Role, int64, error)
	getLookupFn                    func(ctx context.Context, orgID string, tier *rbacdomain.RoleTier, params response.PaginationParams) ([]*rbacdomain.RoleLookup, int64, error)
	createFn                       func(ctx context.Context, role *rbacdomain.Role, permissionIDs []string) error
	updateFn                       func(ctx context.Context, role *rbacdomain.Role, permissionIDs []string) error
	deleteFn                       func(ctx context.Context, id string) error
	getAllRolePermissionsMappingFn func(ctx context.Context) (map[string][]string, error)
	getAllRoleCodesFn              func(ctx context.Context) (map[string]string, error)
}

func (m *mockRoleRepo) GetByID(ctx context.Context, id string) (*rbacdomain.Role, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, rbacdomain.ErrRoleNotFound
}

func (m *mockRoleRepo) GetByIDs(ctx context.Context, ids []string) ([]*rbacdomain.Role, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return nil, nil
}

func (m *mockRoleRepo) GetByCode(ctx context.Context, orgID *string, code string) (*rbacdomain.Role, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, orgID, code)
	}
	return nil, rbacdomain.ErrRoleNotFound
}

func (m *mockRoleRepo) List(ctx context.Context, filter rbacdomain.RoleFilter, params response.PaginationParams) ([]*rbacdomain.Role, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, filter, params)
	}
	return nil, 0, nil
}

func (m *mockRoleRepo) GetLookup(ctx context.Context, orgID string, tier *rbacdomain.RoleTier, params response.PaginationParams) ([]*rbacdomain.RoleLookup, int64, error) {
	if m.getLookupFn != nil {
		return m.getLookupFn(ctx, orgID, tier, params)
	}
	return nil, 0, nil
}

func (m *mockRoleRepo) Create(ctx context.Context, role *rbacdomain.Role, permissionIDs []string) error {
	if m.createFn != nil {
		return m.createFn(ctx, role, permissionIDs)
	}
	return nil
}

func (m *mockRoleRepo) Update(ctx context.Context, role *rbacdomain.Role, permissionIDs []string) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, role, permissionIDs)
	}
	return nil
}

func (m *mockRoleRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRoleRepo) GetAllRolePermissionsMapping(ctx context.Context) (map[string][]string, error) {
	if m.getAllRolePermissionsMappingFn != nil {
		return m.getAllRolePermissionsMappingFn(ctx)
	}
	return nil, nil
}

func (m *mockRoleRepo) GetAllRoleCodes(ctx context.Context) (map[string]string, error) {
	if m.getAllRoleCodesFn != nil {
		return m.getAllRoleCodesFn(ctx)
	}
	return nil, nil
}

func TestUserUsecase_List(t *testing.T) {
	orgID := uuid.NewString()
	brandID := uuid.NewString()
	outletID := uuid.NewString()
	roleID := uuid.NewString()

	users := []*userdomain.User{
		{
			ID:             uuid.NewString(),
			OrganizationID: orgID,
			BrandID:        &brandID,
			OutletID:       &outletID,
			RoleID:         &roleID,
			Username:       "kasir1",
			Email:          "kasir1@harmoni.id",
			Status:         "ACTIVE",
		},
		{
			ID:             uuid.NewString(),
			OrganizationID: orgID,
			Username:       "owner",
			Email:          "owner@harmoni.id",
			Status:         "ACTIVE",
		},
	}

	userRepo := &mockUserRepo{
		listFn: func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.User, int64, error) {
			return users, int64(len(users)), nil
		},
	}

	orgRepo := &mockOrgRepo{
		getByIDsFn: func(ctx context.Context, ids []string) ([]*orgdomain.Organization, error) {
			assert.Contains(t, ids, orgID)
			return []*orgdomain.Organization{
				{ID: orgID, Name: "PT Harmoni", Slug: "pt-harmoni"},
			}, nil
		},
	}

	brandRepo := &mockBrandRepo{
		getByIDsFn: func(ctx context.Context, ids []string) ([]*orgdomain.Brand, error) {
			assert.Contains(t, ids, brandID)
			return []*orgdomain.Brand{
				{ID: brandID, Name: "Daily Coffee", Slug: "daily-coffee"},
			}, nil
		},
	}

	outletRepo := &mockOutletRepo{
		getByIDsFn: func(ctx context.Context, ids []string) ([]*orgdomain.Outlet, error) {
			assert.Contains(t, ids, outletID)
			return []*orgdomain.Outlet{
				{ID: outletID, Name: "Kemang", Code: "DC-KMG", Address: "Jl Kemang"},
			}, nil
		},
	}

	roleRepo := &mockRoleRepo{
		getByIDsFn: func(ctx context.Context, ids []string) ([]*rbacdomain.Role, error) {
			assert.Contains(t, ids, roleID)
			return []*rbacdomain.Role{
				{ID: roleID, Code: "CASHIER", Name: "Kasir", Tier: rbacdomain.RoleTierOutlet},
			}, nil
		},
	}

	uc := usecase.NewUserUsecase(userRepo, orgRepo, brandRepo, outletRepo, roleRepo, zap.NewNop())
	res, total, err := uc.List(context.Background(), userdomain.UserFilter{OrganizationID: orgID}, response.PaginationParams{Page: 1, PerPage: 10})

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, res, 2)

	// Verify relation injection
	assert.Equal(t, "PT Harmoni", res[0].Organization.Name)
	assert.Equal(t, "Daily Coffee", res[0].Brand.Name)
	assert.Equal(t, "Kemang", res[0].Outlet.Name)
	assert.NotNil(t, res[0].Role)
	assert.Equal(t, "CASHIER", res[0].Role.Code)

	assert.Equal(t, "PT Harmoni", res[1].Organization.Name)
	assert.Nil(t, res[1].Brand)
	assert.Nil(t, res[1].Outlet)
	assert.Nil(t, res[1].Role)
}

func TestUserUsecase_Create(t *testing.T) {
	orgID := uuid.NewString()
	brandID := uuid.NewString()
	outletID := uuid.NewString()

	t.Run("success create user with hashed password and PIN", func(t *testing.T) {
		var createdUser *userdomain.User
		userRepo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, orgID, username string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
			createFn: func(ctx context.Context, user *userdomain.User) error {
				createdUser = user
				return nil
			},
		}

		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID, Name: "PT Harmoni"}, nil
			},
		}

		brandRepo := &mockBrandRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Brand, error) {
				return &orgdomain.Brand{ID: brandID, OrganizationID: orgID, Name: "Daily"}, nil
			},
		}

		outletRepo := &mockOutletRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Outlet, error) {
				return &orgdomain.Outlet{ID: outletID, OrganizationID: orgID, BrandID: brandID, Name: "Kemang"}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, brandRepo, outletRepo, &mockRoleRepo{}, zap.NewNop())
		input := usecase.CreateUserInput{
			OrganizationID: orgID,
			BrandID:        &brandID,
			OutletID:       &outletID,
			Username:       "kasir.kemang",
			Email:          "kasir@harmoni.id",
			Password:       "Password123!",
			PIN:            ptr.Ptr("123456"),
			Status:         "ACTIVE",
		}

		res, err := uc.Create(context.Background(), input)
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "kasir.kemang", res.Username)
		assert.Equal(t, "kasir@harmoni.id", res.Email)

		// Assert password and PIN were hashed securely
		require.NotNil(t, createdUser)
		assert.NoError(t, password.Compare(createdUser.Password, "Password123!"))
		require.NotNil(t, createdUser.Pin)
		assert.NoError(t, password.Compare(*createdUser.Pin, "123456"))
	})

	t.Run("conflict duplicate email in organization", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return &userdomain.User{ID: uuid.NewString(), OrganizationID: orgID, Email: email}, nil
			},
		}
		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		_, err := uc.Create(context.Background(), usecase.CreateUserInput{
			OrganizationID: orgID,
			Username:       "newuser",
			Email:          "existing@harmoni.id",
			Password:       "Password123!",
		})

		assert.ErrorIs(t, err, userdomain.ErrEmailAlreadyExists)
	})

	t.Run("hierarchy mismatch when brand belongs to another organization", func(t *testing.T) {
		otherOrgID := uuid.NewString()
		userRepo := &mockUserRepo{}
		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID}, nil
			},
		}
		brandRepo := &mockBrandRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Brand, error) {
				return &orgdomain.Brand{ID: brandID, OrganizationID: otherOrgID}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, brandRepo, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		_, err := uc.Create(context.Background(), usecase.CreateUserInput{
			OrganizationID: orgID,
			BrandID:        &brandID,
			Username:       "newuser",
			Email:          "user@harmoni.id",
			Password:       "Password123!",
		})

		assert.ErrorIs(t, err, userdomain.ErrInvalidHierarchy)
	})

	t.Run("reject assigning DEVELOPER role", func(t *testing.T) {
		devRoleID := uuid.NewString()
		userRepo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, orgID, username string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
		}
		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID}, nil
			},
		}
		roleRepo := &mockRoleRepo{
			getByIDFn: func(ctx context.Context, id string) (*rbacdomain.Role, error) {
				return &rbacdomain.Role{ID: devRoleID, Code: "DEVELOPER"}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, &mockBrandRepo{}, &mockOutletRepo{}, roleRepo, zap.NewNop())
		_, err := uc.Create(context.Background(), usecase.CreateUserInput{
			OrganizationID: orgID,
			RoleID:         &devRoleID,
			Username:       "newdev",
			Email:          "dev@harmoni.id",
			Password:       "Password123!",
		})

		assert.ErrorIs(t, err, rbacdomain.ErrDeveloperRoleProtected)
	})

	t.Run("role not found returns ErrRoleNotFound", func(t *testing.T) {
		invalidRoleID := uuid.NewString()
		userRepo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, orgID, username string) (*userdomain.User, error) {
				return nil, userdomain.ErrUserNotFound
			},
		}
		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID}, nil
			},
		}
		roleRepo := &mockRoleRepo{
			getByIDFn: func(ctx context.Context, id string) (*rbacdomain.Role, error) {
				return nil, rbacdomain.ErrRoleNotFound
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, &mockBrandRepo{}, &mockOutletRepo{}, roleRepo, zap.NewNop())
		_, err := uc.Create(context.Background(), usecase.CreateUserInput{
			OrganizationID: orgID,
			RoleID:         &invalidRoleID,
			Username:       "user1",
			Email:          "user1@harmoni.id",
			Password:       "Password123!",
		})

		assert.ErrorIs(t, err, rbacdomain.ErrRoleNotFound)
	})
}

func TestUserUsecase_GetByID(t *testing.T) {
	orgID := uuid.NewString()
	userID := uuid.NewString()

	t.Run("success get by ID within same organization", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
				return &userdomain.User{
					ID:             userID,
					OrganizationID: orgID,
					Username:       "staff1",
					Email:          "staff1@harmoni.id",
					CreatedAt:      time.Now(),
				}, nil
			},
		}
		orgRepo := &mockOrgRepo{
			getByIDFn: func(ctx context.Context, id string) (*orgdomain.Organization, error) {
				return &orgdomain.Organization{ID: orgID, Name: "PT Harmoni"}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, orgRepo, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		res, err := uc.GetByID(context.Background(), orgID, userID)

		require.NoError(t, err)
		assert.Equal(t, userID, res.ID)
		assert.Equal(t, "staff1", res.Username)
		assert.Equal(t, "PT Harmoni", res.Organization.Name)
	})

	t.Run("tenant violation cross-organization access returns not found", func(t *testing.T) {
		otherOrgID := uuid.NewString()
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
				return &userdomain.User{
					ID:             userID,
					OrganizationID: otherOrgID,
				}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, &mockOrgRepo{}, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		_, err := uc.GetByID(context.Background(), orgID, userID)

		assert.ErrorIs(t, err, userdomain.ErrUserNotFound)
	})
}

func TestUserUsecase_Delete(t *testing.T) {
	orgID := uuid.NewString()
	userID := uuid.NewString()

	t.Run("success delete user", func(t *testing.T) {
		var deletedID string
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
				return &userdomain.User{ID: userID, OrganizationID: orgID}, nil
			},
			deleteFn: func(ctx context.Context, id string) error {
				deletedID = id
				return nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, &mockOrgRepo{}, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		err := uc.Delete(context.Background(), orgID, userID)

		require.NoError(t, err)
		assert.Equal(t, userID, deletedID)
	})

	t.Run("tenant scoping prevents delete of user from different organization", func(t *testing.T) {
		otherOrgID := uuid.NewString()
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id string) (*userdomain.User, error) {
				return &userdomain.User{ID: userID, OrganizationID: otherOrgID}, nil
			},
		}

		uc := usecase.NewUserUsecase(userRepo, &mockOrgRepo{}, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
		err := uc.Delete(context.Background(), orgID, userID)

		assert.ErrorIs(t, err, userdomain.ErrUserNotFound)
	})
}

func TestUserUsecase_GetLookup(t *testing.T) {
	orgID := uuid.NewString()
	outletID := uuid.NewString()

	expectedLookups := []*userdomain.UserLookup{
		{ID: "u-1", Name: "barista.kemang", OutletID: &outletID},
		{ID: "u-2", Name: "kasir.kemang", OutletID: &outletID},
	}

	userRepo := &mockUserRepo{
		getLookupFn: func(ctx context.Context, filter userdomain.UserFilter, params response.PaginationParams) ([]*userdomain.UserLookup, int64, error) {
			assert.Equal(t, orgID, filter.OrganizationID)
			assert.Equal(t, &outletID, filter.OutletID)
			assert.Equal(t, 1, params.Page)
			assert.Equal(t, 20, params.PerPage)
			return expectedLookups, 2, nil
		},
	}

	uc := usecase.NewUserUsecase(userRepo, &mockOrgRepo{}, &mockBrandRepo{}, &mockOutletRepo{}, &mockRoleRepo{}, zap.NewNop())
	res, total, err := uc.GetLookup(context.Background(), userdomain.UserFilter{
		OrganizationID: orgID,
		OutletID:       &outletID,
	}, response.PaginationParams{Page: 1, PerPage: 20})

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, res, 2)
	assert.Equal(t, "barista.kemang", res[0].Name)
	assert.Equal(t, "kasir.kemang", res[1].Name)
}
