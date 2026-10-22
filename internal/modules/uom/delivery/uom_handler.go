package delivery

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	"github.com/hanjo-erp/hanjo/internal/modules/uom/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/uom/usecase"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// UoMHandler exposes Fiber HTTP handlers for UoM management and conversions.
type UoMHandler struct {
	uc usecase.UoMUsecase
}

// NewUoMHandler creates a new UoMHandler instance.
func NewUoMHandler(uc usecase.UoMUsecase) *UoMHandler {
	return &UoMHandler{uc: uc}
}

// RegisterUoMRoutes mounts UoM endpoints under the provided router group.
func RegisterUoMRoutes(router fiber.Router, uc usecase.UoMUsecase, jwtManager *jwt.Manager) *UoMHandler {
	h := NewUoMHandler(uc)

	// UoMs endpoints
	uomsGroup := router.Group("/uoms", middleware.Auth(jwtManager))
	uomsGroup.Get("", h.List)
	// CRITICAL FIBER ROUTING RULE: Register /lookup BEFORE /:id
	uomsGroup.Get("/lookup", h.GetLookup)
	uomsGroup.Get("/:id", h.GetByID)
	uomsGroup.Post("", h.Create)
	uomsGroup.Put("/:id", h.Update)
	uomsGroup.Delete("/:id", h.Delete)

	// UoM Conversions endpoints
	conversionsGroup := router.Group("/uom-conversions", middleware.Auth(jwtManager))
	conversionsGroup.Get("", h.ListConversions)
	conversionsGroup.Post("", h.CreateConversion)
	conversionsGroup.Post("/convert", h.ConvertQuantity)
	conversionsGroup.Delete("/:id", h.DeleteConversion)

	return h
}

// List handles listing UoMs with pagination and filtering.
func (h *UoMHandler) List(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	filter := domain.UoMFilter{
		Search: c.Query("search"),
	}

	categoryQuery := strings.ToUpper(strings.TrimSpace(c.Query("category")))
	if categoryQuery != "" {
		cat := domain.UoMCategory(categoryQuery)
		if cat.IsValid() {
			filter.Category = &cat
		}
	}

	statusQuery := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if statusQuery != "" {
		filter.Status = &statusQuery
	}

	pagination := response.ParsePagination(c)
	uoms, total, err := h.uc.ListUoMs(c.Context(), orgID, filter, pagination)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list uoms")
	}

	meta := response.NewMeta(pagination.Page, pagination.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "uoms retrieved successfully", uoms, meta)
}

// GetLookup handles lightweight, high-speed active UoM retrieval for UI comboboxes.
func (h *UoMHandler) GetLookup(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	var category *domain.UoMCategory
	categoryQuery := strings.ToUpper(strings.TrimSpace(c.Query("category")))
	if categoryQuery != "" {
		cat := domain.UoMCategory(categoryQuery)
		if cat.IsValid() {
			category = &cat
		}
	}

	items, err := h.uc.ListUoMLookup(c.Context(), orgID, category)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch uom lookup")
	}

	return response.Success(c, fiber.StatusOK, "uom lookup retrieved successfully", items)
}

// GetByID retrieves a single UoM by its UUID.
func (h *UoMHandler) GetByID(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	id := c.Params("id")

	uom, err := h.uc.GetUoMByID(c.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrUoMNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "UOM_NOT_FOUND", "unit of measurement not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to retrieve uom")
	}

	return response.Success(c, fiber.StatusOK, "uom retrieved successfully", uom)
}

// Create handles creating a new custom UoM for an organization.
func (h *UoMHandler) Create(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	var req usecase.CreateUoMRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	created, err := h.uc.CreateUoM(c.Context(), orgID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrDuplicateSymbol):
			return response.Fail(c, fiber.StatusConflict, "DUPLICATE_SYMBOL", "a unit with this symbol already exists")
		case errors.Is(err, domain.ErrInvalidUoMCategory):
			return response.Fail(c, fiber.StatusBadRequest, "INVALID_CATEGORY", "invalid measurement category")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create uom")
		}
	}

	return response.Success(c, fiber.StatusCreated, "uom created successfully", created)
}

// Update handles updating a custom UoM.
func (h *UoMHandler) Update(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	id := c.Params("id")

	var req usecase.UpdateUoMRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	updated, err := h.uc.UpdateUoM(c.Context(), orgID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUoMNotFound):
			return response.Fail(c, fiber.StatusNotFound, "UOM_NOT_FOUND", "unit of measurement not found")
		case errors.Is(err, domain.ErrSystemUoMImmutable):
			return response.Fail(c, fiber.StatusForbidden, "SYSTEM_UOM_IMMUTABLE", "system default units cannot be modified")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to update uom")
		}
	}

	return response.Success(c, fiber.StatusOK, "uom updated successfully", updated)
}

// Delete handles soft-deleting a custom UoM.
func (h *UoMHandler) Delete(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	id := c.Params("id")

	if err := h.uc.DeleteUoM(c.Context(), orgID, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrUoMNotFound):
			return response.Fail(c, fiber.StatusNotFound, "UOM_NOT_FOUND", "unit of measurement not found")
		case errors.Is(err, domain.ErrSystemUoMImmutable):
			return response.Fail(c, fiber.StatusForbidden, "SYSTEM_UOM_IMMUTABLE", "system default units cannot be deleted")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete uom")
		}
	}

	return response.Success(c, fiber.StatusOK, "uom deleted successfully", nil)
}

// ListConversions handles listing unit conversion rules.
func (h *UoMHandler) ListConversions(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	conversions, err := h.uc.ListConversions(c.Context(), orgID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list conversions")
	}

	return response.Success(c, fiber.StatusOK, "uom conversions retrieved successfully", conversions)
}

// CreateConversion registers a new conversion factor between two compatible units.
func (h *UoMHandler) CreateConversion(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	var req usecase.CreateConversionRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	created, err := h.uc.CreateConversion(c.Context(), orgID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrIdenticalConversionUnits):
			return response.Fail(c, fiber.StatusBadRequest, "IDENTICAL_UNITS", "source and target units must be different")
		case errors.Is(err, domain.ErrInvalidConversionFactor):
			return response.Fail(c, fiber.StatusBadRequest, "INVALID_FACTOR", "conversion factor must be greater than zero")
		case errors.Is(err, domain.ErrCategoryMismatch):
			return response.Fail(c, fiber.StatusBadRequest, "CATEGORY_MISMATCH", "cannot convert between units of different measurement categories")
		case errors.Is(err, domain.ErrDuplicateConversion):
			return response.Fail(c, fiber.StatusConflict, "DUPLICATE_CONVERSION", "conversion between these units already exists")
		case errors.Is(err, domain.ErrUoMNotFound):
			return response.Fail(c, fiber.StatusNotFound, "UOM_NOT_FOUND", "one or both specified units were not found")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create conversion")
		}
	}

	return response.Success(c, fiber.StatusCreated, "conversion created successfully", created)
}

// DeleteConversion removes a custom unit conversion rule.
func (h *UoMHandler) DeleteConversion(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	id := c.Params("id")

	if err := h.uc.DeleteConversion(c.Context(), orgID, id); err != nil {
		switch {
		case errors.Is(err, domain.ErrConversionNotFound):
			return response.Fail(c, fiber.StatusNotFound, "CONVERSION_NOT_FOUND", "conversion rule not found")
		case errors.Is(err, domain.ErrSystemUoMImmutable):
			return response.Fail(c, fiber.StatusForbidden, "SYSTEM_CONVERSION_IMMUTABLE", "system default conversions cannot be deleted")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete conversion")
		}
	}

	return response.Success(c, fiber.StatusOK, "conversion deleted successfully", nil)
}

// ConvertQuantity calculates quantity in target unit at runtime.
func (h *UoMHandler) ConvertQuantity(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	var req usecase.ConvertQuantityRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "failed to parse request body")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	result, err := h.uc.ConvertQuantity(c.Context(), orgID, req)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrConversionNotFound):
			return response.Fail(c, fiber.StatusNotFound, "CONVERSION_NOT_FOUND", "no direct or inverse conversion rule found between these units")
		case errors.Is(err, domain.ErrCategoryMismatch):
			return response.Fail(c, fiber.StatusBadRequest, "CATEGORY_MISMATCH", "units belong to different measurement categories")
		case errors.Is(err, domain.ErrUoMNotFound):
			return response.Fail(c, fiber.StatusNotFound, "UOM_NOT_FOUND", "one or both specified units were not found")
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to calculate conversion")
		}
	}

	return response.Success(c, fiber.StatusOK, "quantity converted successfully", result)
}
