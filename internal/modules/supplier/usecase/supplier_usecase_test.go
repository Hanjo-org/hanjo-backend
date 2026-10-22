package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/supplier/usecase"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type mockSupplierRepository struct {
	getByIDFn    func(ctx context.Context, id string) (*domain.Supplier, error)
	getByIDsFn   func(ctx context.Context, ids []string) ([]*domain.Supplier, error)
	getByCodeFn  func(ctx context.Context, orgID, code string) (*domain.Supplier, error)
	listFn       func(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]*domain.Supplier, int64, error)
	listLookupFn func(ctx context.Context, orgID string) ([]*domain.Supplier, error)
	createFn     func(ctx context.Context, entity *domain.Supplier) error
	updateFn     func(ctx context.Context, entity *domain.Supplier) error
	deleteFn     func(ctx context.Context, id string) error
}

func (m *mockSupplierRepository) GetByID(ctx context.Context, id string) (*domain.Supplier, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrSupplierNotFound
}

func (m *mockSupplierRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Supplier, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return []*domain.Supplier{}, nil
}

func (m *mockSupplierRepository) GetByCode(ctx context.Context, orgID, code string) (*domain.Supplier, error) {
	if m.getByCodeFn != nil {
		return m.getByCodeFn(ctx, orgID, code)
	}
	return nil, domain.ErrSupplierNotFound
}

func (m *mockSupplierRepository) List(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]*domain.Supplier, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, orgID, filter, pagination)
	}
	return []*domain.Supplier{}, 0, nil
}

func (m *mockSupplierRepository) ListLookup(ctx context.Context, orgID string) ([]*domain.Supplier, error) {
	if m.listLookupFn != nil {
		return m.listLookupFn(ctx, orgID)
	}
	return []*domain.Supplier{}, nil
}

func (m *mockSupplierRepository) Create(ctx context.Context, entity *domain.Supplier) error {
	if m.createFn != nil {
		return m.createFn(ctx, entity)
	}
	return nil
}

func (m *mockSupplierRepository) Update(ctx context.Context, entity *domain.Supplier) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	return nil
}

func (m *mockSupplierRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestSupplierUsecase_CreateSupplier(t *testing.T) {
	orgID := "org-123"

	t.Run("success create supplier with code", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByCodeFn: func(ctx context.Context, oID, code string) (*domain.Supplier, error) {
				return nil, domain.ErrSupplierNotFound
			},
			createFn: func(ctx context.Context, entity *domain.Supplier) error {
				assert.Equal(t, orgID, entity.OrganizationID)
				assert.Equal(t, "PT Diamond Cold Storage", entity.Name)
				assert.Equal(t, "SUPP-01", *entity.Code)
				return nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.CreateSupplier(context.Background(), orgID, usecase.CreateSupplierRequest{
			Name:          "PT Diamond Cold Storage",
			Code:          ptr.Ptr("SUPP-01"),
			ContactPerson: ptr.Ptr("Budi"),
			Phone:         ptr.Ptr("08123456789"),
		})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "PT Diamond Cold Storage", res.Name)
		assert.Equal(t, "ACTIVE", res.Status)
	})

	t.Run("duplicate code returns error", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByCodeFn: func(ctx context.Context, oID, code string) (*domain.Supplier, error) {
				return &domain.Supplier{ID: "supp-existing", OrganizationID: oID, Code: &code}, nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.CreateSupplier(context.Background(), orgID, usecase.CreateSupplierRequest{
			Name: "PT Susu Baru",
			Code: ptr.Ptr("SUPP-01"),
		})

		assert.ErrorIs(t, err, domain.ErrDuplicateSupplierCode)
		assert.Nil(t, res)
	})
}

func TestSupplierUsecase_GetSupplierByID(t *testing.T) {
	orgID := "org-123"

	t.Run("success get supplier", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:             id,
					OrganizationID: orgID,
					Name:           "PT Kopi Jaya",
					Status:         "ACTIVE",
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
				}, nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.GetSupplierByID(context.Background(), orgID, "supp-1")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "PT Kopi Jaya", res.Name)
	})

	t.Run("wrong organization returns not found", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:             id,
					OrganizationID: "different-org",
					Name:           "PT Kopi Jaya",
				}, nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.GetSupplierByID(context.Background(), orgID, "supp-1")

		assert.ErrorIs(t, err, domain.ErrSupplierNotFound)
		assert.Nil(t, res)
	})
}

func TestSupplierUsecase_ListSuppliers(t *testing.T) {
	orgID := "org-123"

	repo := &mockSupplierRepository{
		listFn: func(ctx context.Context, oID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]*domain.Supplier, int64, error) {
			return []*domain.Supplier{
				{ID: "supp-1", OrganizationID: oID, Name: "Supplier A", Status: "ACTIVE"},
				{ID: "supp-2", OrganizationID: oID, Name: "Supplier B", Status: "ACTIVE"},
			}, 2, nil
		},
	}

	uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
	items, total, err := uc.ListSuppliers(context.Background(), orgID, domain.SupplierFilter{}, response.PaginationParams{Page: 1, PerPage: 10})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
}

func TestSupplierUsecase_ListLookupSuppliers(t *testing.T) {
	orgID := "org-123"

	repo := &mockSupplierRepository{
		listLookupFn: func(ctx context.Context, oID string) ([]*domain.Supplier, error) {
			return []*domain.Supplier{
				{ID: "supp-1", OrganizationID: oID, Name: "Supplier A", Code: ptr.Ptr("S1"), Status: "ACTIVE"},
			}, nil
		},
	}

	uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
	items, err := uc.ListLookupSuppliers(context.Background(), orgID)

	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.Equal(t, "Supplier A", items[0].Name)
}

func TestSupplierUsecase_UpdateSupplier(t *testing.T) {
	orgID := "org-123"

	t.Run("success update supplier", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:             id,
					OrganizationID: orgID,
					Name:           "Old Name",
					Status:         "ACTIVE",
				}, nil
			},
			updateFn: func(ctx context.Context, entity *domain.Supplier) error {
				assert.Equal(t, "New Name", entity.Name)
				assert.Equal(t, "INACTIVE", entity.Status)
				return nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.UpdateSupplier(context.Background(), orgID, "supp-1", usecase.UpdateSupplierRequest{
			Name:   "New Name",
			Status: "INACTIVE",
		})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "New Name", res.Name)
		assert.Equal(t, "INACTIVE", res.Status)
	})

	t.Run("update duplicate code returns conflict", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:             id,
					OrganizationID: orgID,
					Name:           "My Supplier",
					Code:           ptr.Ptr("OLD-CODE"),
					Status:         "ACTIVE",
				}, nil
			},
			getByCodeFn: func(ctx context.Context, oID, code string) (*domain.Supplier, error) {
				return &domain.Supplier{
					ID:             "other-supp",
					OrganizationID: oID,
					Code:           &code,
				}, nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		res, err := uc.UpdateSupplier(context.Background(), orgID, "supp-1", usecase.UpdateSupplierRequest{
			Name:   "My Supplier",
			Code:   ptr.Ptr("TAKEN-CODE"),
			Status: "ACTIVE",
		})

		assert.ErrorIs(t, err, domain.ErrDuplicateSupplierCode)
		assert.Nil(t, res)
	})
}

func TestSupplierUsecase_DeleteSupplier(t *testing.T) {
	orgID := "org-123"

	t.Run("success delete", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{ID: id, OrganizationID: orgID}, nil
			},
			deleteFn: func(ctx context.Context, id string) error {
				assert.Equal(t, "supp-1", id)
				return nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		err := uc.DeleteSupplier(context.Background(), orgID, "supp-1")
		assert.NoError(t, err)
	})

	t.Run("delete cross-tenant fails", func(t *testing.T) {
		repo := &mockSupplierRepository{
			getByIDFn: func(ctx context.Context, id string) (*domain.Supplier, error) {
				return &domain.Supplier{ID: id, OrganizationID: "other-org"}, nil
			},
		}

		uc := usecase.NewSupplierUsecase(repo, zap.NewNop())
		err := uc.DeleteSupplier(context.Background(), orgID, "supp-1")
		assert.ErrorIs(t, err, domain.ErrSupplierNotFound)
	})
}
