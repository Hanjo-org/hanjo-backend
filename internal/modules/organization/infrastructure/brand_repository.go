package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type brandRepository struct {
	db *gorm.DB
}

// NewBrandRepository creates a GORM-backed BrandRepository.
func NewBrandRepository(db *gorm.DB) domain.BrandRepository {
	return &brandRepository{db: db}
}

func (r *brandRepository) GetByID(ctx context.Context, id string) (*domain.Brand, error) {
	var model BrandModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBrandNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *brandRepository) GetBySlug(ctx context.Context, orgID, slug string) (*domain.Brand, error) {
	var model BrandModel
	err := r.db.WithContext(ctx).First(&model, "organization_id = ? AND slug = ?", orgID, slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrBrandNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *brandRepository) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*domain.Brand, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&BrandModel{}).Where("organization_id = ?", orgID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []BrandModel
	err := query.Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	brands := make([]*domain.Brand, len(models))
	for i := range models {
		brands[i] = models[i].ToDomain()
	}

	return brands, total, nil
}

func (r *brandRepository) Create(ctx context.Context, brand *domain.Brand) error {
	model := FromBrandDomain(brand)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *brandRepository) Update(ctx context.Context, brand *domain.Brand) error {
	model := FromBrandDomain(brand)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *brandRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&BrandModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrBrandNotFound
	}
	return nil
}
