package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type uomRepository struct {
	db *gorm.DB
}

// NewUoMRepository creates a new UoMRepository backed by GORM.
func NewUoMRepository(db *gorm.DB) domain.UoMRepository {
	return &uomRepository{db: db}
}

func (r *uomRepository) GetByID(ctx context.Context, id string) (*domain.UoM, error) {
	var model UoMModel
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUoMNotFound
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *uomRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.UoM, error) {
	if len(ids) == 0 {
		return []*domain.UoM{}, nil
	}

	var models []UoMModel
	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*domain.UoM, len(models))
	for i, m := range models {
		result[i] = m.ToDomain()
	}
	return result, nil
}

func (r *uomRepository) GetBySymbol(ctx context.Context, orgID *string, symbol string) (*domain.UoM, error) {
	var model UoMModel
	query := r.db.WithContext(ctx).Where("symbol = ?", symbol)

	if orgID != nil {
		query = query.Where("(organization_id = ? OR (is_system = true AND organization_id IS NULL))", *orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	err := query.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUoMNotFound
		}
		return nil, err
	}

	return model.ToDomain(), nil
}

func (r *uomRepository) List(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*domain.UoM, int64, error) {
	var models []UoMModel
	var total int64

	// Scoped to organization custom units OR system default units
	query := r.db.WithContext(ctx).Model(&UoMModel{}).
		Where("(organization_id = ? OR (is_system = true AND organization_id IS NULL))", orgID)

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("(name ILIKE ? OR symbol ILIKE ?)", searchPattern, searchPattern)
	}

	if filter.Category != nil {
		query = query.Where("category = ?", string(*filter.Category))
	}

	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("is_system DESC, name ASC").
		Offset(pagination.Offset()).
		Limit(pagination.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.UoM, len(models))
	for i, m := range models {
		result[i] = m.ToDomain()
	}

	return result, total, nil
}

func (r *uomRepository) ListLookup(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*domain.UoM, error) {
	var models []UoMModel

	query := r.db.WithContext(ctx).
		Where("(organization_id = ? OR (is_system = true AND organization_id IS NULL))", orgID).
		Where("status = ?", "ACTIVE")

	if category != nil {
		query = query.Where("category = ?", string(*category))
	}

	err := query.
		Order("is_system DESC, name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]*domain.UoM, len(models))
	for i, m := range models {
		result[i] = m.ToDomain()
	}

	return result, nil
}

func (r *uomRepository) Create(ctx context.Context, entity *domain.UoM) error {
	model := FromUoMDomain(entity)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *uomRepository) Update(ctx context.Context, entity *domain.UoM) error {
	model := FromUoMDomain(entity)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *uomRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&UoMModel{}).Error
}
