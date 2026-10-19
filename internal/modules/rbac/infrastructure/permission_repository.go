package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
)

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new GORM-backed PermissionRepository.
func NewPermissionRepository(db *gorm.DB) domain.PermissionRepository {
	return &permissionRepository{db: db}
}

func (r *permissionRepository) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	var model PermissionModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *permissionRepository) GetByCode(ctx context.Context, code string) (*domain.Permission, error) {
	var model PermissionModel
	err := r.db.WithContext(ctx).First(&model, "code = ?", code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPermissionNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *permissionRepository) GetByCodes(ctx context.Context, codes []string) ([]*domain.Permission, error) {
	if len(codes) == 0 {
		return []*domain.Permission{}, nil
	}

	var models []PermissionModel
	err := r.db.WithContext(ctx).Where("code IN ?", codes).Find(&models).Error
	if err != nil {
		return nil, err
	}

	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = models[i].ToDomain()
	}
	return perms, nil
}

func (r *permissionRepository) List(ctx context.Context, filter domain.PermissionFilter) ([]*domain.Permission, error) {
	query := r.db.WithContext(ctx).Model(&PermissionModel{})

	if filter.Group != nil && *filter.Group != "" {
		query = query.Where(`"group" = ?`, *filter.Group)
	}
	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}

	var models []PermissionModel
	if err := query.Order(`"group" ASC, code ASC`).Find(&models).Error; err != nil {
		return nil, err
	}

	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = models[i].ToDomain()
	}
	return perms, nil
}

func (r *permissionRepository) ListByRoleID(ctx context.Context, roleID string) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Order(`permissions."group" ASC, permissions.code ASC`).
		Find(&models).Error
	if err != nil {
		return nil, err
	}

	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = models[i].ToDomain()
	}
	return perms, nil
}

func (r *permissionRepository) ListByRoleIDs(ctx context.Context, roleIDs []string) (map[string][]*domain.Permission, error) {
	res := make(map[string][]*domain.Permission)
	if len(roleIDs) == 0 {
		return res, nil
	}

	type resultRow struct {
		RoleID string
		PermissionModel
	}

	var rows []resultRow
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("role_permissions.role_id, permissions.*").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id IN ?", roleIDs).
		Order(`permissions."group" ASC, permissions.code ASC`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		perm := row.PermissionModel.ToDomain()
		res[row.RoleID] = append(res[row.RoleID], perm)
	}

	return res, nil
}
