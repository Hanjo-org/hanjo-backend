package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a GORM-backed UserRepository.
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	var models []UserModel
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		return nil, err
	}

	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = models[i].ToDomain()
	}
	return users, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "email = ?", email).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, orgID, username string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "organization_id = ? AND username = ?", orgID, username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *userRepository) GetByOutletAndUsername(ctx context.Context, outletID, username string) (*domain.User, error) {
	var model UserModel
	err := r.db.WithContext(ctx).First(&model, "outlet_id = ? AND username = ?", outletID, username).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *userRepository) List(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&UserModel{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.BrandID != nil && *filter.BrandID != "" {
		query = query.Where("brand_id = ?", *filter.BrandID)
	}
	if filter.OutletID != nil && *filter.OutletID != "" {
		query = query.Where("outlet_id = ?", *filter.OutletID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		query = query.Where("(username ILIKE ? OR email ILIKE ?)", searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []UserModel
	err := query.Order("created_at DESC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = models[i].ToDomain()
	}
	return users, total, nil
}

func (r *userRepository) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*domain.User, int64, error) {
	return r.List(ctx, domain.UserFilter{OrganizationID: orgID}, params)
}

func (r *userRepository) ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*domain.User, int64, error) {
	return r.List(ctx, domain.UserFilter{OutletID: &outletID}, params)
}

func (r *userRepository) GetLookup(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
	query := r.db.WithContext(ctx).Model(&UserModel{})

	if filter.OrganizationID != "" {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.BrandID != nil && *filter.BrandID != "" {
		query = query.Where("brand_id = ?", *filter.BrandID)
	}
	if filter.OutletID != nil && *filter.OutletID != "" {
		query = query.Where("outlet_id = ?", *filter.OutletID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	} else {
		// By default only return ACTIVE users for lookups/comboboxes
		query = query.Where("status = ?", "ACTIVE")
	}
	if filter.Search != nil && *filter.Search != "" {
		searchPattern := "%" + *filter.Search + "%"
		query = query.Where("username ILIKE ?", searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []struct {
		ID       string  `gorm:"column:id"`
		Username string  `gorm:"column:username"`
		OutletID *string `gorm:"column:outlet_id"`
		BrandID  *string `gorm:"column:brand_id"`
	}

	err := query.Select("id, username, outlet_id, brand_id").
		Order("username ASC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&results).Error
	if err != nil {
		return nil, 0, err
	}

	lookups := make([]*domain.UserLookup, len(results))
	for i, row := range results {
		lookups[i] = &domain.UserLookup{
			ID:       row.ID,
			Name:     row.Username,
			OutletID: row.OutletID,
			BrandID:  row.BrandID,
		}
	}

	return lookups, total, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	model := FromUserDomain(user)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	model := FromUserDomain(user)
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&UserModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}
