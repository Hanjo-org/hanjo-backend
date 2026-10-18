package delivery

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	orgdomain "github.com/hanjo-erp/hanjo/internal/modules/organization/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/user/usecase"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// UserHandler handles HTTP transport for user/staff management endpoints.
type UserHandler struct {
	uc usecase.UserUsecase
}

// RegisterUserRoutes mounts protected user endpoints to the Fiber router.
func RegisterUserRoutes(
	router fiber.Router,
	uc usecase.UserUsecase,
	jwtManager *jwt.Manager,
) *UserHandler {
	h := &UserHandler{uc: uc}

	userGroup := router.Group("/users", middleware.Auth(jwtManager))
	userGroup.Get("", h.List)
	userGroup.Get("/lookup", h.GetLookup)
	userGroup.Post("", h.Create)
	userGroup.Get("/:id", h.GetByID)
	userGroup.Put("/:id", h.Update)
	userGroup.Patch("/:id/status", h.UpdateStatus)
	userGroup.Delete("/:id", h.Delete)

	return h
}

// GetLookup retrieves lightweight active user references for UI comboboxes/dropdowns with pagination.
func (h *UserHandler) GetLookup(c fiber.Ctx) error {
	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	params := response.ParsePagination(c)

	var outletID *string
	if o := strings.TrimSpace(c.Query("outlet_id")); o != "" {
		outletID = ptr.Ptr(o)
	}

	var brandID *string
	if b := strings.TrimSpace(c.Query("brand_id")); b != "" {
		brandID = ptr.Ptr(b)
	}

	var search *string
	if s := strings.TrimSpace(c.Query("search")); s != "" {
		search = ptr.Ptr(s)
	}

	filter := domain.UserFilter{
		OrganizationID: orgID,
		BrandID:        brandID,
		OutletID:       outletID,
		Search:         search,
	}

	lookups, total, err := h.uc.GetLookup(c.Context(), filter, params)
	if err != nil {
		return h.mapUserError(c, err)
	}

	meta := response.NewMeta(params.Page, params.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "lookup pengguna berhasil diambil", lookups, meta)
}

// List retrieves paginated users within tenant scope with preloaded organization, brand, and outlet summaries.
func (h *UserHandler) List(c fiber.Ctx) error {
	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	params := response.ParsePagination(c)

	var brandID *string
	if b := strings.TrimSpace(c.Query("brand_id")); b != "" {
		brandID = ptr.Ptr(b)
	}

	var outletID *string
	if o := strings.TrimSpace(c.Query("outlet_id")); o != "" {
		outletID = ptr.Ptr(o)
	}

	var status *string
	if s := strings.TrimSpace(c.Query("status")); s != "" {
		status = ptr.Ptr(strings.ToUpper(s))
	}

	var search *string
	if q := strings.TrimSpace(c.Query("search")); q != "" {
		search = ptr.Ptr(q)
	}

	filter := domain.UserFilter{
		OrganizationID: orgID,
		BrandID:        brandID,
		OutletID:       outletID,
		Status:         status,
		Search:         search,
	}

	users, total, err := h.uc.List(c.Context(), filter, params)
	if err != nil {
		return h.mapUserError(c, err)
	}

	meta := response.NewMeta(params.Page, params.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "daftar pengguna berhasil diambil", users, meta)
}

// GetByID returns detail of a user including relation summaries.
func (h *UserHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "id pengguna tidak valid")
	}

	orgID := h.resolveOrganizationID(c)
	user, err := h.uc.GetByID(c.Context(), orgID, id)
	if err != nil {
		return h.mapUserError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "data pengguna berhasil ditemukan", user)
}

// Create inserts a new user with hashed credentials and relation validation.
func (h *UserHandler) Create(c fiber.Ctx) error {
	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	var req CreateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	input := usecase.CreateUserInput{
		OrganizationID: orgID,
		BrandID:        req.BrandID,
		OutletID:       req.OutletID,
		Username:       req.Username,
		Email:          req.Email,
		Password:       req.Password,
		PIN:            req.PIN,
		Status:         req.Status,
	}

	user, err := h.uc.Create(c.Context(), input)
	if err != nil {
		return h.mapUserError(c, err)
	}

	return response.Success(c, fiber.StatusCreated, "pengguna berhasil dibuat", user)
}

// Update mutates an existing user profile and assignments.
func (h *UserHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "id pengguna tidak valid")
	}

	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	var req UpdateUserRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	input := usecase.UpdateUserInput{
		BrandID:  req.BrandID,
		OutletID: req.OutletID,
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		PIN:      req.PIN,
		Status:   req.Status,
	}

	user, err := h.uc.Update(c.Context(), orgID, id, input)
	if err != nil {
		return h.mapUserError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "data pengguna berhasil diperbarui", user)
}

// UpdateStatus changes status of a user (ACTIVE, INACTIVE, SUSPENDED).
func (h *UserHandler) UpdateStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "id pengguna tidak valid")
	}

	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	var req UpdateUserStatusRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	user, err := h.uc.UpdateStatus(c.Context(), orgID, id, req.Status)
	if err != nil {
		return h.mapUserError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "status pengguna berhasil diubah", user)
}

// Delete soft-deletes a user from the organization.
func (h *UserHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "id pengguna tidak valid")
	}

	orgID := h.resolveOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	if err := h.uc.Delete(c.Context(), orgID, id); err != nil {
		return h.mapUserError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "pengguna berhasil dihapus", nil)
}

func (h *UserHandler) resolveOrganizationID(c fiber.Ctx) string {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		// Fallback for developer/superadmin queries
		orgID = strings.TrimSpace(c.Query("organization_id"))
	}
	return orgID
}

func (h *UserHandler) mapUserError(c fiber.Ctx, err error) error {
	if errors.Is(err, domain.ErrUserNotFound) {
		return response.Fail(c, fiber.StatusNotFound, "USER_NOT_FOUND", "pengguna tidak ditemukan")
	}
	if errors.Is(err, domain.ErrEmailAlreadyExists) {
		return response.Fail(c, fiber.StatusConflict, "EMAIL_ALREADY_EXISTS", "email sudah terdaftar dalam organisasi ini")
	}
	if errors.Is(err, domain.ErrUsernameAlreadyExists) {
		return response.Fail(c, fiber.StatusConflict, "USERNAME_ALREADY_EXISTS", "username sudah digunakan dalam organisasi ini")
	}
	if errors.Is(err, orgdomain.ErrOrganizationNotFound) {
		return response.Fail(c, fiber.StatusBadRequest, "ORGANIZATION_NOT_FOUND", "organisasi tidak ditemukan")
	}
	if errors.Is(err, orgdomain.ErrBrandNotFound) {
		return response.Fail(c, fiber.StatusBadRequest, "BRAND_NOT_FOUND", "brand tidak ditemukan")
	}
	if errors.Is(err, orgdomain.ErrOutletNotFound) {
		return response.Fail(c, fiber.StatusBadRequest, "OUTLET_NOT_FOUND", "outlet tidak ditemukan")
	}
	if errors.Is(err, domain.ErrInvalidHierarchy) {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_HIERARCHY", "brand atau outlet tidak sesuai dengan hierarki organisasi")
	}
	if errors.Is(err, domain.ErrInvalidStatus) {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_STATUS", "status pengguna tidak valid")
	}
	if errors.Is(err, domain.ErrForbidden) {
		return response.Fail(c, fiber.StatusForbidden, "ACCESS_FORBIDDEN", "akses ditolak")
	}

	return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "terjadi kesalahan internal server")
}
