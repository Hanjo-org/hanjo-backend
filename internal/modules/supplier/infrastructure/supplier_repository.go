package infrastructure

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new SupplierRepository instance.
func NewSupplierRepository(db *gorm.DB) domain.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) GetByID(ctx context.Context, id string) (*domain.Supplier, error) {
	var model SupplierModel
	if err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *supplierRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Supplier, error) {
	if len(ids) == 0 {
		return []*domain.Supplier{}, nil
	}

	var models []SupplierModel
	if err := r.db.WithContext(ctx).Find(&models, "id IN ?", ids).Error; err != nil {
		return nil, err
	}

	results := make([]*domain.Supplier, 0, len(models))
	for i := range models {
		results = append(results, models[i].ToDomain())
	}
	return results, nil
}

func (r *supplierRepository) GetByCode(ctx context.Context, orgID, code string) (*domain.Supplier, error) {
	var model SupplierModel
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND code = ?", orgID, code).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrSupplierNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *supplierRepository) List(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]*domain.Supplier, int64, error) {
	query := r.db.WithContext(ctx).Model(&SupplierModel{}).Where("organization_id = ?", orgID)

	if search := strings.TrimSpace(filter.Search); search != "" {
		searchTerm := "%" + search + "%"
		query = query.Where("(name ILIKE ? OR code ILIKE ? OR contact_person ILIKE ? OR phone ILIKE ?)",
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []SupplierModel
	err := query.
		Order("created_at DESC").
		Offset(pagination.Offset()).
		Limit(pagination.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	results := make([]*domain.Supplier, 0, len(models))
	for i := range models {
		results = append(results, models[i].ToDomain())
	}

	return results, total, nil
}

func (r *supplierRepository) ListLookup(ctx context.Context, orgID string) ([]*domain.Supplier, error) {
	var models []SupplierModel
	err := r.db.WithContext(ctx).
		Select("id", "name", "code", "contact_person", "phone", "status").
		Where("organization_id = ? AND status = 'ACTIVE'", orgID).
		Order("name ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	results := make([]*domain.Supplier, 0, len(models))
	for i := range models {
		results = append(results, models[i].ToDomain())
	}
	return results, nil
}

func (r *supplierRepository) Create(ctx context.Context, entity *domain.Supplier) error {
	model := FromSupplierDomain(entity)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *supplierRepository) Update(ctx context.Context, entity *domain.Supplier) error {
	model := FromSupplierDomain(entity)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *supplierRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&SupplierModel{}, "id = ?", id).Error
}
