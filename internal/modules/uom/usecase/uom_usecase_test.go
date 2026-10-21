package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/uom/usecase"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// mockUoMRepository implements domain.UoMRepository
type mockUoMRepository struct {
	getByIDFn     func(ctx context.Context, id string) (*domain.UoM, error)
	getByIDsFn    func(ctx context.Context, ids []string) ([]*domain.UoM, error)
	getBySymbolFn func(ctx context.Context, orgID *string, symbol string) (*domain.UoM, error)
	listFn        func(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*domain.UoM, int64, error)
	listLookupFn  func(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*domain.UoM, error)
	createFn      func(ctx context.Context, entity *domain.UoM) error
	updateFn      func(ctx context.Context, entity *domain.UoM) error
	deleteFn      func(ctx context.Context, id string) error
}

func (m *mockUoMRepository) GetByID(ctx context.Context, id string) (*domain.UoM, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrUoMNotFound
}

func (m *mockUoMRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.UoM, error) {
	if m.getByIDsFn != nil {
		return m.getByIDsFn(ctx, ids)
	}
	return []*domain.UoM{}, nil
}

func (m *mockUoMRepository) GetBySymbol(ctx context.Context, orgID *string, symbol string) (*domain.UoM, error) {
	if m.getBySymbolFn != nil {
		return m.getBySymbolFn(ctx, orgID, symbol)
	}
	return nil, domain.ErrUoMNotFound
}

func (m *mockUoMRepository) List(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*domain.UoM, int64, error) {
	if m.listFn != nil {
		return m.listFn(ctx, orgID, filter, pagination)
	}
	return []*domain.UoM{}, 0, nil
}

func (m *mockUoMRepository) ListLookup(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*domain.UoM, error) {
	if m.listLookupFn != nil {
		return m.listLookupFn(ctx, orgID, category)
	}
	return []*domain.UoM{}, nil
}

func (m *mockUoMRepository) Create(ctx context.Context, entity *domain.UoM) error {
	if m.createFn != nil {
		return m.createFn(ctx, entity)
	}
	return nil
}

func (m *mockUoMRepository) Update(ctx context.Context, entity *domain.UoM) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, entity)
	}
	return nil
}

func (m *mockUoMRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

// mockConversionRepository implements domain.UoMConversionRepository
type mockConversionRepository struct {
	getByIDFn     func(ctx context.Context, id string) (*domain.UoMConversion, error)
	getByPairFn   func(ctx context.Context, fromUoMID, toUoMID string) (*domain.UoMConversion, error)
	listByUoMIDFn func(ctx context.Context, uomID string) ([]*domain.UoMConversion, error)
	listFn        func(ctx context.Context, orgID string) ([]*domain.UoMConversion, error)
	createFn      func(ctx context.Context, conversion *domain.UoMConversion) error
	deleteFn      func(ctx context.Context, id string) error
}

func (m *mockConversionRepository) GetByID(ctx context.Context, id string) (*domain.UoMConversion, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domain.ErrConversionNotFound
}

func (m *mockConversionRepository) GetByPair(ctx context.Context, fromUoMID, toUoMID string) (*domain.UoMConversion, error) {
	if m.getByPairFn != nil {
		return m.getByPairFn(ctx, fromUoMID, toUoMID)
	}
	return nil, domain.ErrConversionNotFound
}

func (m *mockConversionRepository) ListByUoMID(ctx context.Context, uomID string) ([]*domain.UoMConversion, error) {
	if m.listByUoMIDFn != nil {
		return m.listByUoMIDFn(ctx, uomID)
	}
	return []*domain.UoMConversion{}, nil
}

func (m *mockConversionRepository) List(ctx context.Context, orgID string) ([]*domain.UoMConversion, error) {
	if m.listFn != nil {
		return m.listFn(ctx, orgID)
	}
	return []*domain.UoMConversion{}, nil
}

func (m *mockConversionRepository) Create(ctx context.Context, conversion *domain.UoMConversion) error {
	if m.createFn != nil {
		return m.createFn(ctx, conversion)
	}
	return nil
}

func (m *mockConversionRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestUoMUsecase_CreateUoM(t *testing.T) {
	logger := zap.NewNop()
	orgID := "org-uuid-1"

	tests := []struct {
		name        string
		req         usecase.CreateUoMRequest
		setupMock   func(uomMock *mockUoMRepository)
		expectedErr error
	}{
		{
			name: "success: creates custom UoM",
			req: usecase.CreateUoMRequest{
				Name:          "Karton 12",
				Symbol:        "ctn12",
				Category:      "UNIT",
				DecimalPlaces: 0,
			},
			setupMock: func(uomMock *mockUoMRepository) {
				uomMock.getBySymbolFn = func(ctx context.Context, orgID *string, symbol string) (*domain.UoM, error) {
					return nil, domain.ErrUoMNotFound
				}
				uomMock.createFn = func(ctx context.Context, entity *domain.UoM) error {
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "failure: invalid category",
			req: usecase.CreateUoMRequest{
				Name:          "Invalid Category Unit",
				Symbol:        "icu",
				Category:      "INVALID_CAT",
				DecimalPlaces: 0,
			},
			setupMock:   nil,
			expectedErr: domain.ErrInvalidUoMCategory,
		},
		{
			name: "failure: duplicate symbol",
			req: usecase.CreateUoMRequest{
				Name:          "Duplicate Gram",
				Symbol:        "g",
				Category:      "WEIGHT",
				DecimalPlaces: 2,
			},
			setupMock: func(uomMock *mockUoMRepository) {
				uomMock.getBySymbolFn = func(ctx context.Context, orgID *string, symbol string) (*domain.UoM, error) {
					return &domain.UoM{ID: "system-g", Symbol: "g", IsSystem: true}, nil
				}
			},
			expectedErr: domain.ErrDuplicateSymbol,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uomMock := &mockUoMRepository{}
			convMock := &mockConversionRepository{}
			if tt.setupMock != nil {
				tt.setupMock(uomMock)
			}

			uc := usecase.NewUoMUsecase(uomMock, convMock, logger)
			res, err := uc.CreateUoM(context.Background(), orgID, tt.req)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.req.Name, res.Name)
				assert.Equal(t, tt.req.Symbol, res.Symbol)
				assert.False(t, res.IsSystem)
			}
		})
	}
}

func TestUoMUsecase_GetUoMByID(t *testing.T) {
	logger := zap.NewNop()
	orgID := "org-uuid-1"
	otherOrgID := "org-uuid-2"

	tests := []struct {
		name        string
		id          string
		setupMock   func(uomMock *mockUoMRepository)
		expectedErr error
	}{
		{
			name: "success: gets system UoM",
			id:   "system-ml",
			setupMock: func(uomMock *mockUoMRepository) {
				uomMock.getByIDFn = func(ctx context.Context, id string) (*domain.UoM, error) {
					return &domain.UoM{
						ID:       "system-ml",
						Name:     "Milliliter",
						Symbol:   "ml",
						IsSystem: true,
					}, nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "success: gets own organization custom UoM",
			id:   "custom-ctn",
			setupMock: func(uomMock *mockUoMRepository) {
				uomMock.getByIDFn = func(ctx context.Context, id string) (*domain.UoM, error) {
					return &domain.UoM{
						ID:             "custom-ctn",
						OrganizationID: &orgID,
						Name:           "Custom Carton",
						Symbol:         "cctn",
						IsSystem:       false,
					}, nil
				}
			},
			expectedErr: nil,
		},
		{
			name: "failure: custom UoM belongs to another organization",
			id:   "foreign-ctn",
			setupMock: func(uomMock *mockUoMRepository) {
				uomMock.getByIDFn = func(ctx context.Context, id string) (*domain.UoM, error) {
					return &domain.UoM{
						ID:             "foreign-ctn",
						OrganizationID: &otherOrgID,
						Name:           "Other Org Carton",
						Symbol:         "octn",
						IsSystem:       false,
					}, nil
				}
			},
			expectedErr: domain.ErrUoMNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uomMock := &mockUoMRepository{}
			convMock := &mockConversionRepository{}
			if tt.setupMock != nil {
				tt.setupMock(uomMock)
			}

			uc := usecase.NewUoMUsecase(uomMock, convMock, logger)
			res, err := uc.GetUoMByID(context.Background(), orgID, tt.id)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedErr))
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}

func TestUoMUsecase_UpdateAndDeleteUoM_SystemProtection(t *testing.T) {
	logger := zap.NewNop()
	orgID := "org-uuid-1"

	uomMock := &mockUoMRepository{
		getByIDFn: func(ctx context.Context, id string) (*domain.UoM, error) {
			return &domain.UoM{
				ID:       "sys-kg",
				Name:     "Kilogram",
				Symbol:   "kg",
				IsSystem: true,
			}, nil
		},
	}
	convMock := &mockConversionRepository{}

	uc := usecase.NewUoMUsecase(uomMock, convMock, logger)

	// Update system UoM must fail
	_, err := uc.UpdateUoM(context.Background(), orgID, "sys-kg", usecase.UpdateUoMRequest{
		Name:   "Edited Kilogram",
		Status: "ACTIVE",
	})
	assert.ErrorIs(t, err, domain.ErrSystemUoMImmutable)

	// Delete system UoM must fail
	err = uc.DeleteUoM(context.Background(), orgID, "sys-kg")
	assert.ErrorIs(t, err, domain.ErrSystemUoMImmutable)
}

func TestUoMUsecase_ConvertQuantity(t *testing.T) {
	logger := zap.NewNop()
	orgID := "org-uuid-1"

	kgUoM := &domain.UoM{
		ID:       "kg-id",
		Name:     "Kilogram",
		Symbol:   "kg",
		Category: domain.UoMCategoryWeight,
		IsSystem: true,
	}
	gUoM := &domain.UoM{
		ID:       "g-id",
		Name:     "Gram",
		Symbol:   "g",
		Category: domain.UoMCategoryWeight,
		IsSystem: true,
	}
	mlUoM := &domain.UoM{
		ID:       "ml-id",
		Name:     "Milliliter",
		Symbol:   "ml",
		Category: domain.UoMCategoryVolume,
		IsSystem: true,
	}

	tests := []struct {
		name          string
		req           usecase.ConvertQuantityRequest
		setupMock     func(uomMock *mockUoMRepository, convMock *mockConversionRepository)
		expectedQty   float64
		expectedError error
	}{
		{
			name: "success: direct conversion kg -> g (factor 1000)",
			req: usecase.ConvertQuantityRequest{
				FromUoMID: "kg-id",
				ToUoMID:   "g-id",
				Quantity:  2.5,
			},
			setupMock: func(uomMock *mockUoMRepository, convMock *mockConversionRepository) {
				uomMock.getByIDsFn = func(ctx context.Context, ids []string) ([]*domain.UoM, error) {
					return []*domain.UoM{kgUoM, gUoM}, nil
				}
				convMock.getByPairFn = func(ctx context.Context, from, to string) (*domain.UoMConversion, error) {
					if from == "kg-id" && to == "g-id" {
						return &domain.UoMConversion{
							FromUoMID: from,
							ToUoMID:   to,
							Factor:    1000.0,
						}, nil
					}
					return nil, domain.ErrConversionNotFound
				}
			},
			expectedQty:   2500.0,
			expectedError: nil,
		},
		{
			name: "success: inverse conversion g -> kg (factor 1/1000)",
			req: usecase.ConvertQuantityRequest{
				FromUoMID: "g-id",
				ToUoMID:   "kg-id",
				Quantity:  500.0,
			},
			setupMock: func(uomMock *mockUoMRepository, convMock *mockConversionRepository) {
				uomMock.getByIDsFn = func(ctx context.Context, ids []string) ([]*domain.UoM, error) {
					return []*domain.UoM{kgUoM, gUoM}, nil
				}
				convMock.getByPairFn = func(ctx context.Context, from, to string) (*domain.UoMConversion, error) {
					if from == "kg-id" && to == "g-id" {
						return &domain.UoMConversion{
							FromUoMID: from,
							ToUoMID:   to,
							Factor:    1000.0,
						}, nil
					}
					return nil, domain.ErrConversionNotFound
				}
			},
			expectedQty:   0.5,
			expectedError: nil,
		},
		{
			name: "success: identical units returns original quantity",
			req: usecase.ConvertQuantityRequest{
				FromUoMID: "kg-id",
				ToUoMID:   "kg-id",
				Quantity:  10.0,
			},
			setupMock: func(uomMock *mockUoMRepository, convMock *mockConversionRepository) {
				uomMock.getByIDsFn = func(ctx context.Context, ids []string) ([]*domain.UoM, error) {
					return []*domain.UoM{kgUoM}, nil
				}
			},
			expectedQty:   10.0,
			expectedError: nil,
		},
		{
			name: "failure: category mismatch kg (WEIGHT) -> ml (VOLUME)",
			req: usecase.ConvertQuantityRequest{
				FromUoMID: "kg-id",
				ToUoMID:   "ml-id",
				Quantity:  1.0,
			},
			setupMock: func(uomMock *mockUoMRepository, convMock *mockConversionRepository) {
				uomMock.getByIDsFn = func(ctx context.Context, ids []string) ([]*domain.UoM, error) {
					return []*domain.UoM{kgUoM, mlUoM}, nil
				}
			},
			expectedQty:   0,
			expectedError: domain.ErrCategoryMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uomMock := &mockUoMRepository{}
			convMock := &mockConversionRepository{}
			if tt.setupMock != nil {
				tt.setupMock(uomMock, convMock)
			}

			uc := usecase.NewUoMUsecase(uomMock, convMock, logger)
			res, err := uc.ConvertQuantity(context.Background(), orgID, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.expectedQty, res.ConvertedQuantity)
			}
		})
	}
}

func init() {
	_ = time.Now()
}
