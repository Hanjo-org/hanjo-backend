package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	rbacusecase "github.com/hanjo-erp/hanjo/internal/modules/rbac/usecase"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// RequirePermission ensures the authenticated user holds the specified permission.
// Users with role 'DEVELOPER' bypass all checks automatically.
func RequirePermission(rbacUC rbacusecase.RBACUsecase, perm string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. DEVELOPER super-role full bypass
		if strings.EqualFold(contextutil.GetUserRole(c), "DEVELOPER") {
			return c.Next()
		}

		roleID := contextutil.GetUserRoleID(c)
		if roleID == "" {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_PERMISSION", "akses ditolak: peran pengguna tidak ditemukan")
		}

		if !rbacUC.HasPermission(roleID, perm) {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_PERMISSION", "akses ditolak: peran pengguna tidak memiliki izin '"+perm+"'")
		}

		return c.Next()
	}
}

// RequireAnyPermission ensures the authenticated user holds at least one of the specified permissions.
// Users with role 'DEVELOPER' bypass all checks automatically.
func RequireAnyPermission(rbacUC rbacusecase.RBACUsecase, perms ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. DEVELOPER super-role full bypass
		if strings.EqualFold(contextutil.GetUserRole(c), "DEVELOPER") {
			return c.Next()
		}

		roleID := contextutil.GetUserRoleID(c)
		if roleID == "" {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_PERMISSION", "akses ditolak: peran pengguna tidak ditemukan")
		}

		for _, perm := range perms {
			if rbacUC.HasPermission(roleID, perm) {
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_PERMISSION", "akses ditolak: pengguna tidak memiliki izin yang sesuai")
	}
}

// RequirePermissionOrSupervisorPIN allows execution if current user holds the required permission,
// or if an authorized supervisor in the same outlet/organization provides a valid PIN via 'X-Supervisor-PIN' header.
func RequirePermissionOrSupervisorPIN(rbacUC rbacusecase.RBACUsecase, userRepo userdomain.UserRepository, perm string) fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. DEVELOPER super-role bypass
		if strings.EqualFold(contextutil.GetUserRole(c), "DEVELOPER") {
			return c.Next()
		}

		// 2. Direct permission check
		roleID := contextutil.GetUserRoleID(c)
		if roleID != "" && rbacUC.HasPermission(roleID, perm) {
			return c.Next()
		}

		// 3. Supervisor PIN fallback
		supervisorPIN := strings.TrimSpace(c.Get("X-Supervisor-PIN"))
		if supervisorPIN == "" {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_SUPERVISOR_PIN_REQUIRED", "otorisasi supervisor (PIN) diperlukan untuk tindakan ini")
		}

		orgID := contextutil.GetOrganizationID(c)
		outletID := contextutil.GetOutletID(c)
		supervisorUsername := strings.TrimSpace(c.Get("X-Supervisor-Username"))

		// Case A: Specific supervisor username is provided in request
		if supervisorUsername != "" {
			var supervisor *userdomain.User
			var err error
			if outletID != "" {
				supervisor, err = userRepo.GetByOutletAndUsername(c.Context(), outletID, supervisorUsername)
			}
			if err != nil || supervisor == nil {
				if orgID != "" {
					supervisor, err = userRepo.GetByUsername(c.Context(), orgID, supervisorUsername)
				}
			}

			if err == nil && supervisor != nil && strings.EqualFold(supervisor.Status, "ACTIVE") && supervisor.Pin != nil {
				if password.Compare(*supervisor.Pin, supervisorPIN) == nil {
					// Verify supervisor holds the permission
					if supervisor.RoleID != nil && rbacUC.HasPermission(*supervisor.RoleID, perm) {
						contextutil.SetSupervisorID(c, supervisor.ID)
						return c.Next()
					}
				}
			}
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_INVALID_SUPERVISOR_PIN", "PIN supervisor tidak valid atau supervisor tidak memiliki wewenang")
		}

		// Case B: Search active staff in outlet or organization with matching PIN and valid permission
		var outletPtr *string
		if outletID != "" {
			outletPtr = &outletID
		}

		filter := userdomain.UserFilter{
			OrganizationID: orgID,
			OutletID:       outletPtr,
			Status:         ptr.Ptr("ACTIVE"),
		}

		users, _, err := userRepo.List(c.Context(), filter, response.PaginationParams{Page: 1, PerPage: 50})
		if err != nil {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_INVALID_SUPERVISOR_PIN", "gagal memvalidasi otorisasi supervisor")
		}

		for _, u := range users {
			if u.Pin == nil || *u.Pin == "" {
				continue
			}
			if u.RoleID == nil {
				continue
			}
			if !rbacUC.HasPermission(*u.RoleID, perm) {
				continue
			}
			if password.Compare(*u.Pin, supervisorPIN) == nil {
				contextutil.SetSupervisorID(c, u.ID)
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_INVALID_SUPERVISOR_PIN", "PIN supervisor tidak valid atau supervisor tidak memiliki wewenang")
	}
}
