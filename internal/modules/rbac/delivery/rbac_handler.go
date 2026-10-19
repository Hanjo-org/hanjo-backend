package delivery

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/rbac/usecase"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// RBACHandler exposes Fiber HTTP handlers for roles and permissions.
type RBACHandler struct {
	uc usecase.RBACUsecase
}

// NewRBACHandler creates a new RBACHandler instance.
func NewRBACHandler(uc usecase.RBACUsecase) *RBACHandler {
	return &RBACHandler{uc: uc}
}

// RegisterRBACRoutes mounts role and permission endpoints under the provided router group.
func RegisterRBACRoutes(router fiber.Router, uc usecase.RBACUsecase, jwtManager *jwt.Manager) *RBACHandler {
	h := NewRBACHandler(uc)

	// Roles endpoints
	rolesGroup := router.Group("/roles", middleware.Auth(jwtManager))
	rolesGroup.Get("", h.List)
	// CRITICAL FIBER ROUTING RULE: Register /lookup BEFORE /:id
	rolesGroup.Get("/lookup", h.GetLookup)
	rolesGroup.Get("/:id", h.GetByID)
	rolesGroup.Post("", h.Create)
	rolesGroup.Put("/:id", h.Update)
	rolesGroup.Delete("/:id", h.Delete)

	// Permissions endpoints
	permissionsGroup := router.Group("/permissions", middleware.Auth(jwtManager))
	permissionsGroup.Get("/lookup", h.GetPermissionsLookup)

	return h
}

// List godoc
// @Summary      Daftar Peran (Roles List)
// @Description  Mengambil daftar seluruh peran (sistem default dan custom) dalam organisasi dengan pagination dan filter
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tier      query     string  false  "Filter hierarki peran (ORGANIZATION, BRAND, OUTLET)"
// @Param        status    query     string  false  "Filter status peran (ACTIVE, INACTIVE)"
// @Param        search    query     string  false  "Pencarian nama atau kode peran"
// @Param        page      query     int     false  "Nomor halaman (default 1)"   default(1)
// @Param        per_page  query     int     false  "Jumlah item per halaman (default 20, max 100)" default(20)
// @Success      200       {object}  response.Response{data=[]usecase.RoleResponse,meta=response.PaginationMeta}
// @Failure      401       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles [get]
func (h *RBACHandler) List(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	params := response.ParsePagination(c)

	filter := domain.RoleFilter{
		OrganizationID: &orgID,
	}

	if tier := c.Query("tier"); tier != "" {
		t := domain.RoleTier(strings.ToUpper(tier))
		filter.Tier = &t
	}
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}
	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	roles, total, err := h.uc.ListRoles(c.Context(), filter, params)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	meta := response.NewMeta(params.Page, params.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "daftar peran berhasil diambil", roles, meta)
}

// GetByID godoc
// @Summary      Detail Peran & Izin
// @Description  Mengambil informasi detail satu peran berdasarkan ID beserta izin yang dimilikinya
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "UUID Peran"
// @Success      200  {object}  response.Response{data=usecase.RoleResponse}
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /roles/{id} [get]
func (h *RBACHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "ID peran diperlukan")
	}

	role, err := h.uc.GetRoleByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "ROLE_NOT_FOUND", "peran tidak ditemukan")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "detail peran berhasil diambil", role)
}

// Create godoc
// @Summary      Buat Peran Kustom Baru
// @Description  Membuat peran kustom baru untuk organisasi beserta daftar hak akses yang diberikan
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      delivery.CreateRoleRequest  true  "Create Role Payload"
// @Success      201      {object}  response.Response{data=usecase.RoleResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      409      {object}  response.Response
// @Failure      422      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /roles [post]
func (h *RBACHandler) Create(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	if orgID == "" {
		return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
	}

	var req CreateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	input := usecase.CreateRoleInput{
		OrganizationID: orgID,
		Code:           req.Code,
		Name:           req.Name,
		Description:    req.Description,
		Tier:           req.Tier,
		PermissionIDs:  req.PermissionIDs,
	}

	role, err := h.uc.CreateRole(c.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidRoleCode):
			return response.Fail(c, fiber.StatusBadRequest, "INVALID_ROLE_CODE", err.Error())
		case errors.Is(err, domain.ErrDeveloperRoleProtected):
			return response.Fail(c, fiber.StatusForbidden, "ROLE_PROTECTED", err.Error())
		case errors.Is(err, domain.ErrRoleCodeExists):
			return response.Fail(c, fiber.StatusConflict, "ROLE_CODE_EXISTS", err.Error())
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
	}

	return response.Success(c, fiber.StatusCreated, "peran kustom berhasil dibuat", role)
}

// Update godoc
// @Summary      Perbarui Peran Kustom
// @Description  Memperbarui nama, deskripsi, status, atau daftar hak akses peran kustom
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      string                      true  "UUID Peran"
// @Param        request  body      delivery.UpdateRoleRequest  true  "Update Role Payload"
// @Success      200      {object}  response.Response{data=usecase.RoleResponse}
// @Failure      400      {object}  response.Response
// @Failure      401      {object}  response.Response
// @Failure      403      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      422      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /roles/{id} [put]
func (h *RBACHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "ID peran diperlukan")
	}

	var req UpdateRoleRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	input := usecase.UpdateRoleInput{
		Name:          req.Name,
		Description:   req.Description,
		PermissionIDs: req.PermissionIDs,
		Status:        req.Status,
	}

	role, err := h.uc.UpdateRole(c.Context(), id, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrRoleNotFound):
			return response.Fail(c, fiber.StatusNotFound, "ROLE_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrSystemRoleProtected):
			return response.Fail(c, fiber.StatusForbidden, "ROLE_PROTECTED", err.Error())
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
	}

	return response.Success(c, fiber.StatusOK, "peran berhasil diperbarui", role)
}

// Delete godoc
// @Summary      Hapus Peran Kustom (Soft-Delete)
// @Description  Menghapus peran kustom organisasi. Peran sistem bawaan dan peran yang masih dipakai pengguna aktif dilarang dihapus.
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "UUID Peran"
// @Success      200  {object}  response.Response
// @Failure      400  {object}  response.Response
// @Failure      401  {object}  response.Response
// @Failure      403  {object}  response.Response
// @Failure      404  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /roles/{id} [delete]
func (h *RBACHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_ID", "ID peran diperlukan")
	}

	if err := h.uc.DeleteRole(c.Context(), id); err != nil {
		switch {
		case errors.Is(err, domain.ErrRoleNotFound):
			return response.Fail(c, fiber.StatusNotFound, "ROLE_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrSystemRoleProtected):
			return response.Fail(c, fiber.StatusForbidden, "ROLE_PROTECTED", err.Error())
		case errors.Is(err, domain.ErrRoleInUse):
			return response.Fail(c, fiber.StatusConflict, "ROLE_IN_USE", err.Error())
		default:
			return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		}
	}

	return response.Success(c, fiber.StatusOK, "peran berhasil dihapus", nil)
}

// GetLookup godoc
// @Summary      Lookup Peran / Combobox Options
// @Description  Mengambil referensi ringkas peran aktif untuk UI dropdown combobox (DEVELOPER disaring/tersembunyi)
// @Tags         roles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        tier      query     string  false  "Filter tier peran (ORGANIZATION, BRAND, OUTLET)"
// @Param        page      query     int     false  "Nomor halaman (default 1)"   default(1)
// @Param        per_page  query     int     false  "Jumlah item per halaman (default 20, max 100)" default(20)
// @Success      200       {object}  response.Response{data=[]usecase.RoleLookupResponse,meta=response.PaginationMeta}
// @Failure      401       {object}  response.Response
// @Failure      500       {object}  response.Response
// @Router       /roles/lookup [get]
func (h *RBACHandler) GetLookup(c fiber.Ctx) error {
	orgID := contextutil.GetOrganizationID(c)
	params := response.ParsePagination(c)

	var tier *domain.RoleTier
	if t := c.Query("tier"); t != "" {
		parsedTier := domain.RoleTier(strings.ToUpper(t))
		tier = &parsedTier
	}

	lookups, total, err := h.uc.GetRolesLookup(c.Context(), orgID, tier, params)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	meta := response.NewMeta(params.Page, params.PerPage, total)
	return response.SuccessWithMeta(c, fiber.StatusOK, "lookup peran berhasil diambil", lookups, meta)
}

// GetPermissionsLookup godoc
// @Summary      Lookup Katalog Izin (Grouped by Module)
// @Description  Mengambil master katalog seluruh izin yang dikelompokkan berdasarkan domain modul untuk render matriks checkbox UI
// @Tags         permissions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Response{data=[]usecase.PermissionGroupResponse}
// @Failure      401  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /permissions/lookup [get]
func (h *RBACHandler) GetPermissionsLookup(c fiber.Ctx) error {
	groups, err := h.uc.GetPermissionsLookup(c.Context())
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}

	return response.Success(c, fiber.StatusOK, "daftar izin berhasil diambil", groups)
}
