package infrastructure

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new GORM-backed RoleRepository.
func NewRoleRepository(db *gorm.DB) domain.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	var model RoleModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *roleRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Role, error) {
	if len(ids) == 0 {
		return []*domain.Role{}, nil
	}

	var models []RoleModel
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&models).Error
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = models[i].ToDomain()
	}
	return roles, nil
}

func (r *roleRepository) GetByCode(ctx context.Context, orgID *string, code string) (*domain.Role, error) {
	var model RoleModel
	query := r.db.WithContext(ctx).Where("code = ?", code)

	if orgID != nil && *orgID != "" {
		query = query.Where("organization_id = ?", *orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	err := query.First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRoleNotFound
		}
		return nil, err
	}
	return model.ToDomain(), nil
}

func (r *roleRepository) List(ctx context.Context, filter domain.RoleFilter, params response.PaginationParams) ([]*domain.Role, int64, error) {
	query := r.db.WithContext(ctx).Model(&RoleModel{})

	if filter.OrganizationID != nil && *filter.OrganizationID != "" {
		// Include organization-specific custom roles + system-wide default roles
		query = query.Where("(organization_id = ? OR organization_id IS NULL)", *filter.OrganizationID)
	}
	if filter.Tier != nil {
		query = query.Where("tier = ?", string(*filter.Tier))
	}
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ? OR description ILIKE ?", pattern, pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var models []RoleModel
	err := query.
		Order("is_system DESC, tier ASC, name ASC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = models[i].ToDomain()
	}
	return roles, total, nil
}

// GetLookup retrieves safe, minimal role records for UI comboboxes and dropdowns.
// Crucial: Automatically excludes 'DEVELOPER' role and filters for ACTIVE status!
func (r *roleRepository) GetLookup(ctx context.Context, orgID string, tier *domain.RoleTier, params response.PaginationParams) ([]*domain.RoleLookup, int64, error) {
	query := r.db.WithContext(ctx).Model(&RoleModel{}).
		Where("status = ?", "ACTIVE").
		Where("code != ?", "DEVELOPER") // Never leak DEVELOPER super-role to tenant comboboxes!

	if orgID != "" {
		query = query.Where("(organization_id = ? OR organization_id IS NULL)", orgID)
	} else {
		query = query.Where("organization_id IS NULL")
	}

	if tier != nil {
		query = query.Where("tier = ?", string(*tier))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type lookupRow struct {
		ID   string
		Code string
		Name string
		Tier string
	}

	var rows []lookupRow
	err := query.
		Select("id, code, name, tier").
		Order("is_system DESC, tier ASC, name ASC").
		Offset(params.Offset()).
		Limit(params.PerPage).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	lookups := make([]*domain.RoleLookup, len(rows))
	for i, row := range rows {
		lookups[i] = &domain.RoleLookup{
			ID:   row.ID,
			Code: row.Code,
			Name: row.Name,
			Tier: domain.RoleTier(row.Tier),
		}
	}
	return lookups, total, nil
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role, permissionIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := FromRoleDomain(role)
		if err := tx.Create(model).Error; err != nil {
			return err
		}

		if len(permissionIDs) > 0 {
			now := time.Now().UTC()
			mappings := make([]RolePermissionModel, len(permissionIDs))
			for i, pID := range permissionIDs {
				mappings[i] = RolePermissionModel{
					RoleID:       model.ID,
					PermissionID: pID,
					CreatedAt:    now,
				}
			}
			if err := tx.Create(&mappings).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role, permissionIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := FromRoleDomain(role)
		if err := tx.Save(model).Error; err != nil {
			return err
		}

		// Replace permissions: delete old, insert new
		if err := tx.Where("role_id = ?", model.ID).Delete(&RolePermissionModel{}).Error; err != nil {
			return err
		}

		if len(permissionIDs) > 0 {
			now := time.Now().UTC()
			mappings := make([]RolePermissionModel, len(permissionIDs))
			for i, pID := range permissionIDs {
				mappings[i] = RolePermissionModel{
					RoleID:       model.ID,
					PermissionID: pID,
					CreatedAt:    now,
				}
			}
			if err := tx.Create(&mappings).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *roleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&RoleModel{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrRoleNotFound
	}
	return nil
}

func (r *roleRepository) GetAllRolePermissionsMapping(ctx context.Context) (map[string][]string, error) {
	if r.db == nil {
		return make(map[string][]string), nil
	}

	type permJoinRow struct {
		RoleID         string
		PermissionCode string
	}

	var rows []permJoinRow
	err := r.db.WithContext(ctx).
		Table("role_permissions").
		Select("role_permissions.role_id, permissions.code AS permission_code").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, row := range rows {
		result[row.RoleID] = append(result[row.RoleID], row.PermissionCode)
	}
	return result, nil
}

func (r *roleRepository) GetAllRoleCodes(ctx context.Context) (map[string]string, error) {
	if r.db == nil {
		return make(map[string]string), nil
	}

	type codeRow struct {
		ID   string
		Code string
	}

	var rows []codeRow
	err := r.db.WithContext(ctx).
		Table("roles").
		Select("id, code").
		Where("deleted_at IS NULL").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(rows))
	for _, row := range rows {
		result[row.ID] = row.Code
	}
	return result, nil
}
