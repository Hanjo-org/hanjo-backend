package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type organizationRepository struct {
	db *gorm.DB
}

// NewOrganizationRepository creates a GORM-backed OrganizationRepository.
func NewOrganizationRepository(db *gorm.DB) domain.OrganizationRepository {
	return &organizationRepository{db: db}
}

func (r *organizationRepository) GetByID(ctx context.Context, id string) (*domain.Organization, error) {
	var model OrganizationModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *organizationRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Organization, error) {
	if len(ids) == 0 {
		return []*domain.Organization{}, nil
	}

	var models []OrganizationModel
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		return nil, err
	}

	orgs := make([]*domain.Organization, len(models))
	for i := range models {
		orgs[i] = models[i].ToDomain()
	}
	return orgs, nil
}

func (r *organizationRepository) GetBySlug(ctx context.Context, slug string) (*domain.Organization, error) {
	var model OrganizationModel
	err := r.db.WithContext(ctx).First(&model, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrOrganizationNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *organizationRepository) List(ctx context.Context, params response.PaginationParams) ([]*domain.Organization, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&OrganizationModel{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []OrganizationModel
	err := query.Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	orgs := make([]*domain.Organization, len(models))
	for i := range models {
		orgs[i] = models[i].ToDomain()
	}

	return orgs, total, nil
}

func (r *organizationRepository) Create(ctx context.Context, org *domain.Organization) error {
	model := FromOrganizationDomain(org)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *organizationRepository) Update(ctx context.Context, org *domain.Organization) error {
	model := FromOrganizationDomain(org)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *organizationRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&OrganizationModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrOrganizationNotFound
	}
	return nil
}
