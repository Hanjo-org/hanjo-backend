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

func (r *userRepository) ListByOrganization(ctx context.Context, orgID string, params response.PaginationParams) ([]*domain.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&UserModel{}).Where("organization_id = ?", orgID)

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

func (r *userRepository) ListByOutlet(ctx context.Context, outletID string, params response.PaginationParams) ([]*domain.User, int64, error) {
	var total int64
	query := r.db.WithContext(ctx).Model(&UserModel{}).Where("outlet_id = ?", outletID)

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
