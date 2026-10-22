package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// SupplierUsecase defines application business logic for supplier management.
type SupplierUsecase interface {
	CreateSupplier(ctx context.Context, orgID string, req CreateSupplierRequest) (*SupplierResponse, error)
	GetSupplierByID(ctx context.Context, orgID, id string) (*SupplierResponse, error)
	ListSuppliers(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]SupplierResponse, int64, error)
	ListLookupSuppliers(ctx context.Context, orgID string) ([]SupplierLookupResponse, error)
	UpdateSupplier(ctx context.Context, orgID, id string, req UpdateSupplierRequest) (*SupplierResponse, error)
	DeleteSupplier(ctx context.Context, orgID, id string) error
}

type supplierUsecase struct {
	repo domain.SupplierRepository
	log  *zap.Logger
}

// NewSupplierUsecase initializes a new SupplierUsecase instance.
func NewSupplierUsecase(repo domain.SupplierRepository, log *zap.Logger) SupplierUsecase {
	return &supplierUsecase{
		repo: repo,
		log:  log,
	}
}

func (u *supplierUsecase) CreateSupplier(ctx context.Context, orgID string, req CreateSupplierRequest) (*SupplierResponse, error) {
	name := strings.TrimSpace(req.Name)
	var code *string
	if req.Code != nil {
		trimmedCode := strings.TrimSpace(*req.Code)
		if trimmedCode != "" {
			code = &trimmedCode
			existing, err := u.repo.GetByCode(ctx, orgID, trimmedCode)
			if err != nil && !errors.Is(err, domain.ErrSupplierNotFound) {
				u.log.Error("failed checking supplier code uniqueness", zap.String("org_id", orgID), zap.String("code", trimmedCode), zap.Error(err))
				return nil, err
			}
			if existing != nil {
				return nil, domain.ErrDuplicateSupplierCode
			}
		}
	}

	supplier := domain.NewSupplier(
		orgID,
		name,
		code,
		trimStringPtr(req.ContactPerson),
		trimStringPtr(req.Phone),
		trimStringPtr(req.Email),
		trimStringPtr(req.Address),
	)

	if err := u.repo.Create(ctx, supplier); err != nil {
		u.log.Error("failed creating supplier", zap.String("org_id", orgID), zap.Error(err))
		return nil, err
	}

	res := toSupplierResponse(supplier)
	return &res, nil
}

func (u *supplierUsecase) GetSupplierByID(ctx context.Context, orgID, id string) (*SupplierResponse, error) {
	supplier, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if supplier.OrganizationID != orgID {
		return nil, domain.ErrSupplierNotFound
	}

	res := toSupplierResponse(supplier)
	return &res, nil
}

func (u *supplierUsecase) ListSuppliers(ctx context.Context, orgID string, filter domain.SupplierFilter, pagination response.PaginationParams) ([]SupplierResponse, int64, error) {
	items, total, err := u.repo.List(ctx, orgID, filter, pagination)
	if err != nil {
		u.log.Error("failed listing suppliers", zap.String("org_id", orgID), zap.Error(err))
		return nil, 0, err
	}

	results := make([]SupplierResponse, 0, len(items))
	for _, item := range items {
		results = append(results, toSupplierResponse(item))
	}

	return results, total, nil
}

func (u *supplierUsecase) ListLookupSuppliers(ctx context.Context, orgID string) ([]SupplierLookupResponse, error) {
	items, err := u.repo.ListLookup(ctx, orgID)
	if err != nil {
		u.log.Error("failed listing lookup suppliers", zap.String("org_id", orgID), zap.Error(err))
		return nil, err
	}

	results := make([]SupplierLookupResponse, 0, len(items))
	for _, item := range items {
		results = append(results, SupplierLookupResponse{
			ID:            item.ID,
			Name:          item.Name,
			Code:          item.Code,
			ContactPerson: item.ContactPerson,
			Phone:         item.Phone,
		})
	}

	return results, nil
}

func (u *supplierUsecase) UpdateSupplier(ctx context.Context, orgID, id string, req UpdateSupplierRequest) (*SupplierResponse, error) {
	supplier, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if supplier.OrganizationID != orgID {
		return nil, domain.ErrSupplierNotFound
	}

	var code *string
	if req.Code != nil {
		trimmedCode := strings.TrimSpace(*req.Code)
		if trimmedCode != "" {
			code = &trimmedCode
			if supplier.Code == nil || *supplier.Code != trimmedCode {
				existing, err := u.repo.GetByCode(ctx, orgID, trimmedCode)
				if err != nil && !errors.Is(err, domain.ErrSupplierNotFound) {
					u.log.Error("failed checking supplier code uniqueness on update", zap.String("org_id", orgID), zap.String("code", trimmedCode), zap.Error(err))
					return nil, err
				}
				if existing != nil && existing.ID != supplier.ID {
					return nil, domain.ErrDuplicateSupplierCode
				}
			}
		}
	}

	supplier.Name = strings.TrimSpace(req.Name)
	supplier.Code = code
	supplier.ContactPerson = trimStringPtr(req.ContactPerson)
	supplier.Phone = trimStringPtr(req.Phone)
	supplier.Email = trimStringPtr(req.Email)
	supplier.Address = trimStringPtr(req.Address)
	supplier.Status = req.Status
	supplier.UpdatedAt = time.Now().UTC()

	if err := u.repo.Update(ctx, supplier); err != nil {
		u.log.Error("failed updating supplier", zap.String("id", id), zap.Error(err))
		return nil, err
	}

	res := toSupplierResponse(supplier)
	return &res, nil
}

func (u *supplierUsecase) DeleteSupplier(ctx context.Context, orgID, id string) error {
	supplier, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if supplier.OrganizationID != orgID {
		return domain.ErrSupplierNotFound
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		u.log.Error("failed deleting supplier", zap.String("id", id), zap.Error(err))
		return err
	}

	return nil
}

func toSupplierResponse(s *domain.Supplier) SupplierResponse {
	return SupplierResponse{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Name:           s.Name,
		Code:           s.Code,
		ContactPerson:  s.ContactPerson,
		Phone:          s.Phone,
		Email:          s.Email,
		Address:        s.Address,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}
}

func trimStringPtr(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
