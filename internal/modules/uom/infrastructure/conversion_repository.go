package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
)

type conversionRepository struct {
	db *gorm.DB
}

// NewConversionRepository creates a new UoMConversionRepository backed by GORM.
func NewConversionRepository(db *gorm.DB) domain.UoMConversionRepository {
	return &conversionRepository{db: db}
}

func (r *conversionRepository) GetByID(ctx context.Context, id string) (*domain.UoMConversion, error) {
	var model UoMConversionModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrConversionNotFound
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *conversionRepository) GetByPair(ctx context.Context, fromUoMID, toUoMID string) (*domain.UoMConversion, error) {
	var model UoMConversionModel
	err := r.db.WithContext(ctx).
		Where("from_uom_id = ? AND to_uom_id = ?", fromUoMID, toUoMID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrConversionNotFound
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *conversionRepository) ListByUoMID(ctx context.Context, uomID string) ([]*domain.UoMConversion, error) {
	var models []UoMConversionModel
	err := r.db.WithContext(ctx).
		Where("from_uom_id = ? OR to_uom_id = ?", uomID, uomID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.UoMConversion, len(models))
	for i, m := range models {
		result[i] = m.ToDomain()
	}
	return result, nil
}

func (r *conversionRepository) List(ctx context.Context, orgID string) ([]*domain.UoMConversion, error) {
	var models []UoMConversionModel
	err := r.db.WithContext(ctx).
		Where("organization_id = ? OR (is_system = true AND organization_id IS NULL)", orgID).
		Find(&models).Error

	if err != nil {
		return nil, err
	}

	result := make([]*domain.UoMConversion, len(models))
	for i, m := range models {
		result[i] = m.ToDomain()
	}
	return result, nil
}

func (r *conversionRepository) Create(ctx context.Context, conversion *domain.UoMConversion) error {
	model := FromConversionDomain(conversion)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *conversionRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&UoMConversionModel{}).Error
}
