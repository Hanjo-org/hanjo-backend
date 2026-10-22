package usecase

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// UoMUsecase defines all business operations for UoM and conversions.
type UoMUsecase interface {
	CreateUoM(ctx context.Context, orgID string, req CreateUoMRequest) (*UoMResponse, error)
	GetUoMByID(ctx context.Context, orgID string, id string) (*UoMResponse, error)
	ListUoMs(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*UoMResponse, int64, error)
	ListUoMLookup(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*UoMLookupResponse, error)
	UpdateUoM(ctx context.Context, orgID string, id string, req UpdateUoMRequest) (*UoMResponse, error)
	DeleteUoM(ctx context.Context, orgID string, id string) error

	CreateConversion(ctx context.Context, orgID string, req CreateConversionRequest) (*UoMConversionResponse, error)
	ListConversions(ctx context.Context, orgID string) ([]*UoMConversionResponse, error)
	DeleteConversion(ctx context.Context, orgID string, id string) error
	ConvertQuantity(ctx context.Context, orgID string, req ConvertQuantityRequest) (*ConvertQuantityResponse, error)
}

type uomUsecase struct {
	uomRepo        domain.UoMRepository
	conversionRepo domain.UoMConversionRepository
	log            *zap.Logger
}

// NewUoMUsecase creates a new instance of UoMUsecase.
func NewUoMUsecase(
	uomRepo domain.UoMRepository,
	conversionRepo domain.UoMConversionRepository,
	log *zap.Logger,
) UoMUsecase {
	return &uomUsecase{
		uomRepo:        uomRepo,
		conversionRepo: conversionRepo,
		log:            log,
	}
}

func (uc *uomUsecase) CreateUoM(ctx context.Context, orgID string, req CreateUoMRequest) (*UoMResponse, error) {
	category := domain.UoMCategory(req.Category)
	if !category.IsValid() {
		return nil, domain.ErrInvalidUoMCategory
	}

	// Check if symbol already exists within org or system
	existing, err := uc.uomRepo.GetBySymbol(ctx, &orgID, req.Symbol)
	if err == nil && existing != nil {
		return nil, domain.ErrDuplicateSymbol
	} else if err != nil && !errors.Is(err, domain.ErrUoMNotFound) {
		uc.log.Error("failed checking existing uom symbol", zap.Error(err), zap.String("symbol", req.Symbol))
		return nil, err
	}

	entity := domain.NewCustomUoM(orgID, req.Name, req.Symbol, category, req.DecimalPlaces)
	if err := uc.uomRepo.Create(ctx, entity); err != nil {
		uc.log.Error("failed creating custom uom", zap.Error(err), zap.String("org_id", orgID))
		return nil, err
	}

	return toUoMResponse(entity), nil
}

func (uc *uomUsecase) GetUoMByID(ctx context.Context, orgID string, id string) (*UoMResponse, error) {
	entity, err := uc.uomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify accessibility: must be system unit OR belong to user's organization
	if !entity.IsSystem && (entity.OrganizationID == nil || *entity.OrganizationID != orgID) {
		return nil, domain.ErrUoMNotFound
	}

	return toUoMResponse(entity), nil
}

func (uc *uomUsecase) ListUoMs(ctx context.Context, orgID string, filter domain.UoMFilter, pagination response.PaginationParams) ([]*UoMResponse, int64, error) {
	entities, total, err := uc.uomRepo.List(ctx, orgID, filter, pagination)
	if err != nil {
		uc.log.Error("failed listing uoms", zap.Error(err), zap.String("org_id", orgID))
		return nil, 0, err
	}

	responses := make([]*UoMResponse, len(entities))
	for i, e := range entities {
		responses[i] = toUoMResponse(e)
	}

	return responses, total, nil
}

func (uc *uomUsecase) ListUoMLookup(ctx context.Context, orgID string, category *domain.UoMCategory) ([]*UoMLookupResponse, error) {
	entities, err := uc.uomRepo.ListLookup(ctx, orgID, category)
	if err != nil {
		uc.log.Error("failed fetching uom lookup", zap.Error(err), zap.String("org_id", orgID))
		return nil, err
	}

	responses := make([]*UoMLookupResponse, len(entities))
	for i, e := range entities {
		responses[i] = &UoMLookupResponse{
			ID:            e.ID,
			Name:          e.Name,
			Symbol:        e.Symbol,
			Category:      e.Category,
			DecimalPlaces: e.DecimalPlaces,
		}
	}

	return responses, nil
}

func (uc *uomUsecase) UpdateUoM(ctx context.Context, orgID string, id string, req UpdateUoMRequest) (*UoMResponse, error) {
	entity, err := uc.uomRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if entity.IsSystem {
		return nil, domain.ErrSystemUoMImmutable
	}

	if entity.OrganizationID == nil || *entity.OrganizationID != orgID {
		return nil, domain.ErrUoMNotFound
	}

	entity.Name = req.Name
	entity.Status = req.Status
	entity.UpdatedAt = time.Now().UTC()

	if err := uc.uomRepo.Update(ctx, entity); err != nil {
		uc.log.Error("failed updating uom", zap.Error(err), zap.String("uom_id", id))
		return nil, err
	}

	return toUoMResponse(entity), nil
}

func (uc *uomUsecase) DeleteUoM(ctx context.Context, orgID string, id string) error {
	entity, err := uc.uomRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if entity.IsSystem {
		return domain.ErrSystemUoMImmutable
	}

	if entity.OrganizationID == nil || *entity.OrganizationID != orgID {
		return domain.ErrUoMNotFound
	}

	if err := uc.uomRepo.Delete(ctx, id); err != nil {
		uc.log.Error("failed deleting uom", zap.Error(err), zap.String("uom_id", id))
		return err
	}

	return nil
}

func (uc *uomUsecase) CreateConversion(ctx context.Context, orgID string, req CreateConversionRequest) (*UoMConversionResponse, error) {
	if req.FromUoMID == req.ToUoMID {
		return nil, domain.ErrIdenticalConversionUnits
	}
	if req.Factor <= 0 {
		return nil, domain.ErrInvalidConversionFactor
	}

	// Fetch both units to ensure they exist and share the same measurement category
	uoms, err := uc.uomRepo.GetByIDs(ctx, []string{req.FromUoMID, req.ToUoMID})
	if err != nil {
		return nil, err
	}
	if len(uoms) != 2 {
		return nil, domain.ErrUoMNotFound
	}

	var fromUoM, toUoM *domain.UoM
	for _, u := range uoms {
		if u.ID == req.FromUoMID {
			fromUoM = u
		}
		if u.ID == req.ToUoMID {
			toUoM = u
		}
	}

	if fromUoM == nil || toUoM == nil {
		return nil, domain.ErrUoMNotFound
	}

	if fromUoM.Category != toUoM.Category {
		return nil, domain.ErrCategoryMismatch
	}

	// Verify accessibility of both units
	if (!fromUoM.IsSystem && (fromUoM.OrganizationID == nil || *fromUoM.OrganizationID != orgID)) ||
		(!toUoM.IsSystem && (toUoM.OrganizationID == nil || *toUoM.OrganizationID != orgID)) {
		return nil, domain.ErrUoMNotFound
	}

	// Check for existing conversion pair
	existing, err := uc.conversionRepo.GetByPair(ctx, req.FromUoMID, req.ToUoMID)
	if err == nil && existing != nil {
		return nil, domain.ErrDuplicateConversion
	}

	conversion := domain.NewConversion(&orgID, req.FromUoMID, req.ToUoMID, req.Factor, false)
	if err := uc.conversionRepo.Create(ctx, conversion); err != nil {
		uc.log.Error("failed creating uom conversion", zap.Error(err))
		return nil, err
	}

	return &UoMConversionResponse{
		ID:             conversion.ID,
		OrganizationID: conversion.OrganizationID,
		FromUoMID:      fromUoM.ID,
		FromUoMName:    fromUoM.Name,
		FromUoMSymbol:  fromUoM.Symbol,
		ToUoMID:        toUoM.ID,
		ToUoMName:      toUoM.Name,
		ToUoMSymbol:    toUoM.Symbol,
		Factor:         conversion.Factor,
		IsSystem:       conversion.IsSystem,
		CreatedAt:      conversion.CreatedAt,
	}, nil
}

func (uc *uomUsecase) ListConversions(ctx context.Context, orgID string) ([]*UoMConversionResponse, error) {
	conversions, err := uc.conversionRepo.List(ctx, orgID)
	if err != nil {
		uc.log.Error("failed listing uom conversions", zap.Error(err))
		return nil, err
	}

	// Batch fetch referenced units
	unitIDMap := make(map[string]bool)
	for _, c := range conversions {
		unitIDMap[c.FromUoMID] = true
		unitIDMap[c.ToUoMID] = true
	}

	unitIDs := make([]string, 0, len(unitIDMap))
	for id := range unitIDMap {
		unitIDs = append(unitIDs, id)
	}

	uoms, err := uc.uomRepo.GetByIDs(ctx, unitIDs)
	if err != nil {
		return nil, err
	}

	uomMap := make(map[string]*domain.UoM, len(uoms))
	for _, u := range uoms {
		uomMap[u.ID] = u
	}

	responses := make([]*UoMConversionResponse, len(conversions))
	for i, c := range conversions {
		res := &UoMConversionResponse{
			ID:             c.ID,
			OrganizationID: c.OrganizationID,
			FromUoMID:      c.FromUoMID,
			ToUoMID:        c.ToUoMID,
			Factor:         c.Factor,
			IsSystem:       c.IsSystem,
			CreatedAt:      c.CreatedAt,
		}
		if from, ok := uomMap[c.FromUoMID]; ok {
			res.FromUoMName = from.Name
			res.FromUoMSymbol = from.Symbol
		}
		if to, ok := uomMap[c.ToUoMID]; ok {
			res.ToUoMName = to.Name
			res.ToUoMSymbol = to.Symbol
		}
		responses[i] = res
	}

	return responses, nil
}

func (uc *uomUsecase) DeleteConversion(ctx context.Context, orgID string, id string) error {
	conversion, err := uc.conversionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if conversion.IsSystem {
		return domain.ErrSystemUoMImmutable
	}

	if conversion.OrganizationID == nil || *conversion.OrganizationID != orgID {
		return domain.ErrConversionNotFound
	}

	return uc.conversionRepo.Delete(ctx, id)
}

func (uc *uomUsecase) ConvertQuantity(ctx context.Context, orgID string, req ConvertQuantityRequest) (*ConvertQuantityResponse, error) {
	if req.Quantity <= 0 {
		return nil, domain.ErrInvalidConversionFactor
	}

	uoms, err := uc.uomRepo.GetByIDs(ctx, []string{req.FromUoMID, req.ToUoMID})
	if err != nil {
		return nil, err
	}
	if len(uoms) != 2 && req.FromUoMID != req.ToUoMID {
		return nil, domain.ErrUoMNotFound
	}

	var fromUoM, toUoM *domain.UoM
	for _, u := range uoms {
		if u.ID == req.FromUoMID {
			fromUoM = u
		}
		if u.ID == req.ToUoMID {
			toUoM = u
		}
	}

	if req.FromUoMID == req.ToUoMID {
		fromUoM = uoms[0]
		toUoM = uoms[0]
		return &ConvertQuantityResponse{
			FromUoM:           fromUoM.ToSummary(),
			ToUoM:             toUoM.ToSummary(),
			OriginalQuantity:  req.Quantity,
			ConvertedQuantity: req.Quantity,
			Factor:            1.0,
		}, nil
	}

	if fromUoM == nil || toUoM == nil {
		return nil, domain.ErrUoMNotFound
	}

	if fromUoM.Category != toUoM.Category {
		return nil, domain.ErrCategoryMismatch
	}

	// 1. Check direct conversion: from -> to
	conv, err := uc.conversionRepo.GetByPair(ctx, req.FromUoMID, req.ToUoMID)
	if err == nil && conv != nil {
		return &ConvertQuantityResponse{
			FromUoM:           fromUoM.ToSummary(),
			ToUoM:             toUoM.ToSummary(),
			OriginalQuantity:  req.Quantity,
			ConvertedQuantity: req.Quantity * conv.Factor,
			Factor:            conv.Factor,
		}, nil
	}

	// 2. Check inverse conversion: to -> from
	invConv, err := uc.conversionRepo.GetByPair(ctx, req.ToUoMID, req.FromUoMID)
	if err == nil && invConv != nil && invConv.Factor > 0 {
		factor := 1.0 / invConv.Factor
		return &ConvertQuantityResponse{
			FromUoM:           fromUoM.ToSummary(),
			ToUoM:             toUoM.ToSummary(),
			OriginalQuantity:  req.Quantity,
			ConvertedQuantity: req.Quantity * factor,
			Factor:            factor,
		}, nil
	}

	return nil, domain.ErrConversionNotFound
}

func toUoMResponse(e *domain.UoM) *UoMResponse {
	return &UoMResponse{
		ID:             e.ID,
		OrganizationID: e.OrganizationID,
		Name:           e.Name,
		Symbol:         e.Symbol,
		Category:       e.Category,
		DecimalPlaces:  e.DecimalPlaces,
		IsSystem:       e.IsSystem,
		Status:         e.Status,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
