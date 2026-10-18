package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	orgdomain "github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// UserUsecase orchestrates user lifecycle operations, multi-tenant scoping, and cross-module batch fetching.
type UserUsecase interface {
	List(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserWithRelations, int64, error)
	GetLookup(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error)
	GetByID(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error)
	Create(ctx context.Context, input CreateUserInput) (*domain.UserWithRelations, error)
	Update(ctx context.Context, orgID, id string, input UpdateUserInput) (*domain.UserWithRelations, error)
	UpdateStatus(ctx context.Context, orgID, id string, status string) (*domain.UserWithRelations, error)
	Delete(ctx context.Context, orgID, id string) error
}

type userUsecase struct {
	userRepo   domain.UserRepository
	orgRepo    orgdomain.OrganizationRepository
	brandRepo  orgdomain.BrandRepository
	outletRepo orgdomain.OutletRepository
	log        *zap.Logger
}

// NewUserUsecase creates a new UserUsecase instance with cross-module repository dependencies.
func NewUserUsecase(
	userRepo domain.UserRepository,
	orgRepo orgdomain.OrganizationRepository,
	brandRepo orgdomain.BrandRepository,
	outletRepo orgdomain.OutletRepository,
	log *zap.Logger,
) UserUsecase {
	return &userUsecase{
		userRepo:   userRepo,
		orgRepo:    orgRepo,
		brandRepo:  brandRepo,
		outletRepo: outletRepo,
		log:        log,
	}
}

// List retrieves paginated users with dynamically batched relations to prevent N+1 queries.
func (u *userUsecase) List(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserWithRelations, int64, error) {
	users, total, err := u.userRepo.List(ctx, filter, params)
	if err != nil {
		u.log.Error("failed to list users", zap.Error(err), zap.String("org_id", filter.OrganizationID))
		return nil, 0, err
	}

	if len(users) == 0 {
		return []*domain.UserWithRelations{}, total, nil
	}

	// 1. Collect unique IDs for cross-module batch fetching (In-Memory deduplication)
	orgIDSet := make(map[string]struct{})
	brandIDSet := make(map[string]struct{})
	outletIDSet := make(map[string]struct{})

	for _, user := range users {
		if user.OrganizationID != "" {
			orgIDSet[user.OrganizationID] = struct{}{}
		}
		if user.BrandID != nil && *user.BrandID != "" {
			brandIDSet[*user.BrandID] = struct{}{}
		}
		if user.OutletID != nil && *user.OutletID != "" {
			outletIDSet[*user.OutletID] = struct{}{}
		}
	}

	orgIDs := make([]string, 0, len(orgIDSet))
	for id := range orgIDSet {
		orgIDs = append(orgIDs, id)
	}

	brandIDs := make([]string, 0, len(brandIDSet))
	for id := range brandIDSet {
		brandIDs = append(brandIDs, id)
	}

	outletIDs := make([]string, 0, len(outletIDSet))
	for id := range outletIDSet {
		outletIDs = append(outletIDs, id)
	}

	// 2. Batch-fetch organizations (Single query, No N+1)
	orgMap := make(map[string]*domain.OrganizationSummary, len(orgIDs))
	if len(orgIDs) > 0 {
		orgs, err := u.orgRepo.GetByIDs(ctx, orgIDs)
		if err != nil {
			u.log.Warn("failed to batch fetch organizations for users", zap.Error(err))
		} else {
			for _, org := range orgs {
				orgMap[org.ID] = &domain.OrganizationSummary{
					ID:   org.ID,
					Name: org.Name,
					Slug: org.Slug,
				}
			}
		}
	}

	// 3. Batch-fetch brands (Single query, No N+1)
	brandMap := make(map[string]*domain.BrandSummary, len(brandIDs))
	if len(brandIDs) > 0 {
		brands, err := u.brandRepo.GetByIDs(ctx, brandIDs)
		if err != nil {
			u.log.Warn("failed to batch fetch brands for users", zap.Error(err))
		} else {
			for _, b := range brands {
				brandMap[b.ID] = &domain.BrandSummary{
					ID:      b.ID,
					Name:    b.Name,
					Slug:    b.Slug,
					LogoURL: b.LogoURL,
				}
			}
		}
	}

	// 4. Batch-fetch outlets (Single query, No N+1)
	outletMap := make(map[string]*domain.OutletSummary, len(outletIDs))
	if len(outletIDs) > 0 {
		outlets, err := u.outletRepo.GetByIDs(ctx, outletIDs)
		if err != nil {
			u.log.Warn("failed to batch fetch outlets for users", zap.Error(err))
		} else {
			for _, o := range outlets {
				outletMap[o.ID] = &domain.OutletSummary{
					ID:      o.ID,
					Name:    o.Name,
					Code:    o.Code,
					Address: o.Address,
				}
			}
		}
	}

	// 5. Assemble composite UserWithRelations in memory
	result := make([]*domain.UserWithRelations, len(users))
	for i, user := range users {
		var orgSummary *domain.OrganizationSummary
		if s, ok := orgMap[user.OrganizationID]; ok {
			orgSummary = s
		}

		var brandSummary *domain.BrandSummary
		if user.BrandID != nil {
			if s, ok := brandMap[*user.BrandID]; ok {
				brandSummary = s
			}
		}

		var outletSummary *domain.OutletSummary
		if user.OutletID != nil {
			if s, ok := outletMap[*user.OutletID]; ok {
				outletSummary = s
			}
		}

		result[i] = &domain.UserWithRelations{
			User:         user,
			Organization: orgSummary,
			Brand:        brandSummary,
			Outlet:       outletSummary,
		}
	}

	return result, total, nil
}

// GetLookup returns lightweight active user references for dropdown comboboxes with pagination.
func (u *userUsecase) GetLookup(ctx context.Context, filter domain.UserFilter, params response.PaginationParams) ([]*domain.UserLookup, int64, error) {
	lookups, total, err := u.userRepo.GetLookup(ctx, filter, params)
	if err != nil {
		u.log.Error("failed to get user lookups", zap.Error(err), zap.String("org_id", filter.OrganizationID))
		return nil, 0, err
	}
	return lookups, total, nil
}

// GetByID retrieves a single user scoped to tenant with relations populated.
func (u *userUsecase) GetByID(ctx context.Context, orgID, id string) (*domain.UserWithRelations, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Tenant boundary isolation
	if orgID != "" && user.OrganizationID != orgID {
		return nil, domain.ErrUserNotFound
	}

	return u.populateRelations(ctx, user)
}

// Create validates, hashes credentials, and persists a new user entity.
func (u *userUsecase) Create(ctx context.Context, input CreateUserInput) (*domain.UserWithRelations, error) {
	// 1. Validate Organization existence
	if _, err := u.orgRepo.GetByID(ctx, input.OrganizationID); err != nil {
		return nil, orgdomain.ErrOrganizationNotFound
	}

	// 2. Validate Brand if specified
	if input.BrandID != nil && *input.BrandID != "" {
		brand, err := u.brandRepo.GetByID(ctx, *input.BrandID)
		if err != nil {
			return nil, orgdomain.ErrBrandNotFound
		}
		if brand.OrganizationID != input.OrganizationID {
			return nil, domain.ErrInvalidHierarchy
		}
	}

	// 3. Validate Outlet if specified
	if input.OutletID != nil && *input.OutletID != "" {
		outlet, err := u.outletRepo.GetByID(ctx, *input.OutletID)
		if err != nil {
			return nil, orgdomain.ErrOutletNotFound
		}
		if outlet.OrganizationID != input.OrganizationID {
			return nil, domain.ErrInvalidHierarchy
		}
		if input.BrandID != nil && *input.BrandID != "" && outlet.BrandID != *input.BrandID {
			return nil, domain.ErrInvalidHierarchy
		}
	}

	// 4. Check uniqueness of email within organization
	existingEmail, err := u.userRepo.GetByEmail(ctx, input.Email)
	if err == nil && existingEmail != nil && existingEmail.OrganizationID == input.OrganizationID {
		return nil, domain.ErrEmailAlreadyExists
	}

	// 5. Check uniqueness of username within organization
	existingUsername, err := u.userRepo.GetByUsername(ctx, input.OrganizationID, input.Username)
	if err == nil && existingUsername != nil {
		return nil, domain.ErrUsernameAlreadyExists
	}

	// 6. Hash password securely
	hashedPassword, err := password.Hash(input.Password)
	if err != nil {
		u.log.Error("failed to hash user password", zap.Error(err))
		return nil, err
	}

	// 7. Hash PIN if provided
	var hashedPIN *string
	if input.PIN != nil && *input.PIN != "" {
		hPin, err := password.Hash(*input.PIN)
		if err != nil {
			u.log.Error("failed to hash user PIN", zap.Error(err))
			return nil, err
		}
		hashedPIN = &hPin
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	if status == "" {
		status = "ACTIVE"
	}
	if status != "ACTIVE" && status != "INACTIVE" && status != "SUSPENDED" {
		return nil, domain.ErrInvalidStatus
	}

	now := time.Now().UTC()
	newUser := &domain.User{
		ID:             uuid.NewString(),
		OrganizationID: input.OrganizationID,
		BrandID:        input.BrandID,
		OutletID:       input.OutletID,
		Username:       input.Username,
		Email:          input.Email,
		Password:       hashedPassword,
		Pin:            hashedPIN,
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := u.userRepo.Create(ctx, newUser); err != nil {
		u.log.Error("failed to create user", zap.Error(err))
		return nil, err
	}

	return u.populateRelations(ctx, newUser)
}

// Update mutates existing user attributes and re-verifies uniqueness if email/username changed.
func (u *userUsecase) Update(ctx context.Context, orgID, id string, input UpdateUserInput) (*domain.UserWithRelations, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if orgID != "" && user.OrganizationID != orgID {
		return nil, domain.ErrUserNotFound
	}

	// Brand hierarchy check
	if input.BrandID != nil {
		if *input.BrandID != "" {
			brand, err := u.brandRepo.GetByID(ctx, *input.BrandID)
			if err != nil {
				return nil, orgdomain.ErrBrandNotFound
			}
			if brand.OrganizationID != user.OrganizationID {
				return nil, domain.ErrInvalidHierarchy
			}
			user.BrandID = input.BrandID
		} else {
			user.BrandID = nil
		}
	}

	// Outlet hierarchy check
	if input.OutletID != nil {
		if *input.OutletID != "" {
			outlet, err := u.outletRepo.GetByID(ctx, *input.OutletID)
			if err != nil {
				return nil, orgdomain.ErrOutletNotFound
			}
			if outlet.OrganizationID != user.OrganizationID {
				return nil, domain.ErrInvalidHierarchy
			}
			if user.BrandID != nil && *user.BrandID != "" && outlet.BrandID != *user.BrandID {
				return nil, domain.ErrInvalidHierarchy
			}
			user.OutletID = input.OutletID
		} else {
			user.OutletID = nil
		}
	}

	// Email update
	if input.Email != nil && *input.Email != "" && !strings.EqualFold(*input.Email, user.Email) {
		existingEmail, err := u.userRepo.GetByEmail(ctx, *input.Email)
		if err == nil && existingEmail != nil && existingEmail.OrganizationID == user.OrganizationID && existingEmail.ID != user.ID {
			return nil, domain.ErrEmailAlreadyExists
		}
		user.Email = *input.Email
	}

	// Username update
	if input.Username != nil && *input.Username != "" && !strings.EqualFold(*input.Username, user.Username) {
		existingUsername, err := u.userRepo.GetByUsername(ctx, user.OrganizationID, *input.Username)
		if err == nil && existingUsername != nil && existingUsername.ID != user.ID {
			return nil, domain.ErrUsernameAlreadyExists
		}
		user.Username = *input.Username
	}

	// Password update
	if input.Password != nil && *input.Password != "" {
		hashed, err := password.Hash(*input.Password)
		if err != nil {
			return nil, err
		}
		user.Password = hashed
	}

	// PIN update
	if input.PIN != nil {
		if *input.PIN != "" {
			hashedPIN, err := password.Hash(*input.PIN)
			if err != nil {
				return nil, err
			}
			user.Pin = &hashedPIN
		} else {
			user.Pin = nil
		}
	}

	// Status update
	if input.Status != nil && *input.Status != "" {
		st := strings.ToUpper(strings.TrimSpace(*input.Status))
		if st != "ACTIVE" && st != "INACTIVE" && st != "SUSPENDED" {
			return nil, domain.ErrInvalidStatus
		}
		user.Status = st
	}

	user.UpdatedAt = time.Now().UTC()
	if err := u.userRepo.Update(ctx, user); err != nil {
		u.log.Error("failed to update user", zap.Error(err), zap.String("user_id", id))
		return nil, err
	}

	return u.populateRelations(ctx, user)
}

// UpdateStatus changes active status of a user (ACTIVE, INACTIVE, SUSPENDED).
func (u *userUsecase) UpdateStatus(ctx context.Context, orgID, id string, status string) (*domain.UserWithRelations, error) {
	st := strings.ToUpper(strings.TrimSpace(status))
	if st != "ACTIVE" && st != "INACTIVE" && st != "SUSPENDED" {
		return nil, domain.ErrInvalidStatus
	}

	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if orgID != "" && user.OrganizationID != orgID {
		return nil, domain.ErrUserNotFound
	}

	user.Status = st
	user.UpdatedAt = time.Now().UTC()
	if err := u.userRepo.Update(ctx, user); err != nil {
		u.log.Error("failed to update user status", zap.Error(err), zap.String("user_id", id))
		return nil, err
	}

	return u.populateRelations(ctx, user)
}

// Delete marks a user as soft-deleted.
func (u *userUsecase) Delete(ctx context.Context, orgID, id string) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if orgID != "" && user.OrganizationID != orgID {
		return domain.ErrUserNotFound
	}

	return u.userRepo.Delete(ctx, id)
}

func (u *userUsecase) populateRelations(ctx context.Context, user *domain.User) (*domain.UserWithRelations, error) {
	var orgSummary *domain.OrganizationSummary
	if user.OrganizationID != "" {
		org, err := u.orgRepo.GetByID(ctx, user.OrganizationID)
		if err == nil && org != nil {
			orgSummary = &domain.OrganizationSummary{
				ID:   org.ID,
				Name: org.Name,
				Slug: org.Slug,
			}
		}
	}

	var brandSummary *domain.BrandSummary
	if user.BrandID != nil && *user.BrandID != "" {
		brand, err := u.brandRepo.GetByID(ctx, *user.BrandID)
		if err == nil && brand != nil {
			brandSummary = &domain.BrandSummary{
				ID:      brand.ID,
				Name:    brand.Name,
				Slug:    brand.Slug,
				LogoURL: brand.LogoURL,
			}
		}
	}

	var outletSummary *domain.OutletSummary
	if user.OutletID != nil && *user.OutletID != "" {
		outlet, err := u.outletRepo.GetByID(ctx, *user.OutletID)
		if err == nil && outlet != nil {
			outletSummary = &domain.OutletSummary{
				ID:      outlet.ID,
				Name:    outlet.Name,
				Code:    outlet.Code,
				Address: outlet.Address,
			}
		}
	}

	return &domain.UserWithRelations{
		User:         user,
		Organization: orgSummary,
		Brand:        brandSummary,
		Outlet:       outletSummary,
	}, nil
}
