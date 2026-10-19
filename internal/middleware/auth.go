package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

// Auth guards route groups requiring a valid RS256 Bearer token.
func Auth(manager *jwt.Manager) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)
		if header == "" {
			return response.Fail(c, fiber.StatusUnauthorized, "AUTH_UNAUTHORIZED", "header otorisasi tidak ditemukan")
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return response.Fail(c, fiber.StatusUnauthorized, "AUTH_UNAUTHORIZED", "format header otorisasi harus Bearer <token>")
		}

		claims, err := manager.ParseToken(parts[1])
		if err != nil {
			return response.Fail(c, fiber.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "token tidak valid atau telah kedaluwarsa")
		}

		contextutil.SetUserID(c, claims.UserID)
		contextutil.SetUserRole(c, claims.Role)
		if claims.RoleID != "" {
			contextutil.SetUserRoleID(c, claims.RoleID)
		}
		if claims.Email != "" {
			contextutil.SetUserEmail(c, claims.Email)
		}
		if claims.OrganizationID != "" {
			contextutil.SetOrganizationID(c, claims.OrganizationID)
		}
		if claims.BrandID != "" {
			contextutil.SetBrandID(c, claims.BrandID)
		}
		if claims.OutletID != "" {
			contextutil.SetOutletID(c, claims.OutletID)
		}
		c.Locals("claims", claims)

		return c.Next()
	}
}

// RequireRoles restricts endpoint access to users possessing one of the allowed roles.
func RequireRoles(allowedRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole := contextutil.GetUserRole(c)
		if userRole == "" {
			return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN", "akses ditolak: peran pengguna tidak ditemukan")
		}

		for _, role := range allowedRoles {
			if strings.EqualFold(role, userRole) {
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "AUTH_FORBIDDEN_ROLE", "akses ditolak: peran pengguna tidak memiliki izin")
	}
}

// RequireOrganization ensures the request contains an active organization context.
func RequireOrganization() fiber.Handler {
	return func(c fiber.Ctx) error {
		orgID := contextutil.GetOrganizationID(c)
		if orgID == "" {
			return response.Fail(c, fiber.StatusBadRequest, "TENANT_REQUIRED", "konteks organisasi diperlukan")
		}
		return c.Next()
	}
}
