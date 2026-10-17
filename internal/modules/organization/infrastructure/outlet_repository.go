package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type outletRepository struct {
	db *gorm.DB
}

// NewOutletRepository creates a GORM-backed OutletRepository.
func NewOutletRepository(db *gorm.DB) domain.OutletRepository {
	return &outletRepository{db: db}
}

func (r *outletRepository) GetByID(ctx context.Context, id string) (*domain.Outlet, error) {
	var model OutletModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOutletNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *outletRepository) GetByCode(ctx context.Context, orgID, code string) (*domain.Outlet, error) {
	var model OutletModel
	err := r.db.WithContext(ctx).First(&model, "organization_id = ? AND code = ?", orgID, code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOutletNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *outletRepository) ListByBrand(ctx context.Context, brandID string, params response.PaginationParams) ([]*domain.Outlet, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&OutletModel{}).Where("brand_id = ?", brandID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []OutletModel
	err := query.Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	outlets := make([]*domain.Outlet, len(models))
	for i := range models {
		outlets[i] = models[i].ToDomain()
	}

	return outlets, total, nil
}

func (r *outletRepository) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*domain.Outlet, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&OutletModel{}).Where("organization_id = ?", orgID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []OutletModel
	err := query.Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	outlets := make([]*domain.Outlet, len(models))
	for i := range models {
		outlets[i] = models[i].ToDomain()
	}

	return outlets, total, nil
}

func (r *outletRepository) Create(ctx context.Context, outlet *domain.Outlet) error {
	model := FromOutletDomain(outlet)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *outletRepository) Update(ctx context.Context, outlet *domain.Outlet) error {
	model := FromOutletDomain(outlet)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *outletRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&OutletModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOutletNotFound
	}
	return nil
}
