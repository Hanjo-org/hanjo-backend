package delivery

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	"github.com/hanjo-erp/hanjo/internal/modules/supplier/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/supplier/usecase"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// SupplierHandler exposes Fiber HTTP handlers for supplier management.
type SupplierHandler struct {
	uc usecase.SupplierUsecase
}

// NewSupplierHandler creates a new SupplierHandler instance.
func NewSupplierHandler(uc usecase.SupplierUsecase) *SupplierHandler {
	return &SupplierHandler{uc: uc}
}

// RegisterSupplierRoutes mounts supplier endpoints under the provided router group.
func RegisterSupplierRoutes(router fiber.Router, uc usecase.SupplierUsecase, jwtManager *jwt.Manager) *SupplierHandler {
	h := NewSupplierHandler(uc)

	group := router.Group("/suppliers", middleware.Auth(jwtManager))
	group.Get("", h.List)
	// CRITICAL FIBER ROUTING RULE: Register /lookup BEFORE /:id
	group.Get("/lookup", h.GetLookup)
	group.Get("/:id", h.GetByID)
	group.Post("", h.Create)
	group.Put("/:id", h.Update)
	group.Delete("/:id", h.Delete)

	return h
}

// List handles listing suppliers with pagination, search, and status filtering.
func (h *SupplierHandler) List(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	filter := domain.SupplierFilter{
		Search: c.Query("search"),
	}

	statusQuery := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	if statusQuery != "" {
		filter.Status = &statusQuery
	}

	pagination := response.ParsePagination(c)
	suppliers, total, err := h.uc.ListSuppliers(c.Context(), orgID, filter, pagination)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list suppliers")
	}

	meta := response.NewMeta(pagination.Page, pagination.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "suppliers retrieved successfully", suppliers, meta)
}

// GetLookup handles lightweight, high-speed active supplier retrieval for UI comboboxes.
func (h *SupplierHandler) GetLookup(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	lookups, err := h.uc.ListLookupSuppliers(c.Context(), orgID)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list lookup suppliers")
	}

	return response.Success(c, fiber.StatusOK, "supplier lookup retrieved successfully", lookups)
}

// GetByID handles retrieving a single supplier by its unique identifier.
func (h *SupplierHandler) GetByID(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "supplier ID is required")
	}

	res, err := h.uc.GetSupplierByID(c.Context(), orgID, id)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "NOT_FOUND", "supplier not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to get supplier")
	}

	return response.Success(c, fiber.StatusOK, "supplier retrieved successfully", res)
}

// Create handles creating a new supplier.
func (h *SupplierHandler) Create(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	var req usecase.CreateSupplierRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed request payload")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	res, err := h.uc.CreateSupplier(c.Context(), orgID, req)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateSupplierCode) {
			return response.Fail(c, fiber.StatusConflict, "DUPLICATE_CODE", "supplier code already exists in this organization")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create supplier")
	}

	return response.Success(c, fiber.StatusCreated, "supplier created successfully", res)
}

// Update handles modifying an existing supplier.
func (h *SupplierHandler) Update(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "supplier ID is required")
	}

	var req usecase.UpdateSupplierRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "malformed request payload")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	res, err := h.uc.UpdateSupplier(c.Context(), orgID, id, req)
	if err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "NOT_FOUND", "supplier not found")
		}
		if errors.Is(err, domain.ErrDuplicateSupplierCode) {
			return response.Fail(c, fiber.StatusConflict, "DUPLICATE_CODE", "supplier code already exists in this organization")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to update supplier")
	}

	return response.Success(c, fiber.StatusOK, "supplier updated successfully", res)
}

// Delete handles soft-deleting an existing supplier.
func (h *SupplierHandler) Delete(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing organization context")
	}

	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "supplier ID is required")
	}

	if err := h.uc.DeleteSupplier(c.Context(), orgID, id); err != nil {
		if errors.Is(err, domain.ErrSupplierNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "NOT_FOUND", "supplier not found")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete supplier")
	}

	return response.Success(c, fiber.StatusOK, "supplier deleted successfully", nil)
}
