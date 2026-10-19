package contextutil

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
)

var (
	ErrUserUnauthenticated = errors.New("user is unauthenticated")
)

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	UserEmailKey      contextKey = "user_email"
	UserRoleKey       contextKey = "user_role"
	UserRoleIDKey     contextKey = "user_role_id"
	OrganizationIDKey contextKey = "organization_id"
	BrandIDKey        contextKey = "brand_id"
	OutletIDKey       contextKey = "outlet_id"
	SupervisorIDKey   contextKey = "supervisor_id"
)

// SetUserID sets the authenticated user ID in fiber.Ctx locals.
func SetUserID(c fiber.Ctx, userID string) {
	c.Locals(string(UserIDKey), userID)
}

// GetUserID retrieves the authenticated user ID from fiber.Ctx locals.
// Returns ErrUserUnauthenticated if missing or empty.
func GetUserID(c fiber.Ctx) (string, error) {
	val, ok := c.Locals(string(UserIDKey)).(string)
	if !ok || val == "" {
		return "", ErrUserUnauthenticated
	}
	return val, nil
}

// MustGetUserID retrieves the user ID from fiber.Ctx locals, returning empty string if not found.
func MustGetUserID(c fiber.Ctx) string {
	val, _ := GetUserID(c)
	return val
}

// SetUserEmail sets the authenticated user email in fiber.Ctx locals.
func SetUserEmail(c fiber.Ctx, email string) {
	c.Locals(string(UserEmailKey), email)
}

// GetUserEmail retrieves the authenticated user email from fiber.Ctx locals.
func GetUserEmail(c fiber.Ctx) string {
	val, ok := c.Locals(string(UserEmailKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetUserRole sets the authenticated user role in fiber.Ctx locals.
func SetUserRole(c fiber.Ctx, role string) {
	c.Locals(string(UserRoleKey), role)
}

// GetUserRole retrieves the authenticated user role from fiber.Ctx locals.
func GetUserRole(c fiber.Ctx) string {
	val, ok := c.Locals(string(UserRoleKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetUserRoleID sets the authenticated user role ID in fiber.Ctx locals.
func SetUserRoleID(c fiber.Ctx, roleID string) {
	c.Locals(string(UserRoleIDKey), roleID)
}

// GetUserRoleID retrieves the authenticated user role ID from fiber.Ctx locals.
func GetUserRoleID(c fiber.Ctx) string {
	val, ok := c.Locals(string(UserRoleIDKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetOrganizationID sets the authenticated organization ID in fiber.Ctx locals.
func SetOrganizationID(c fiber.Ctx, orgID string) {
	c.Locals(string(OrganizationIDKey), orgID)
}

// GetOrganizationID retrieves the organization ID from fiber.Ctx locals.
func GetOrganizationID(c fiber.Ctx) string {
	val, ok := c.Locals(string(OrganizationIDKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetBrandID sets the active brand ID in fiber.Ctx locals.
func SetBrandID(c fiber.Ctx, brandID string) {
	c.Locals(string(BrandIDKey), brandID)
}

// GetBrandID retrieves the active brand ID from fiber.Ctx locals.
func GetBrandID(c fiber.Ctx) string {
	val, ok := c.Locals(string(BrandIDKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetOutletID sets the active outlet ID in fiber.Ctx locals.
func SetOutletID(c fiber.Ctx, outletID string) {
	c.Locals(string(OutletIDKey), outletID)
}

// GetOutletID retrieves the active outlet ID from fiber.Ctx locals.
func GetOutletID(c fiber.Ctx) string {
	val, ok := c.Locals(string(OutletIDKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// SetSupervisorID sets the authorized supervisor user ID in fiber.Ctx locals.
func SetSupervisorID(c fiber.Ctx, supervisorID string) {
	c.Locals(string(SupervisorIDKey), supervisorID)
}

// GetSupervisorID retrieves the authorized supervisor user ID from fiber.Ctx locals.
func GetSupervisorID(c fiber.Ctx) string {
	val, ok := c.Locals(string(SupervisorIDKey)).(string)
	if !ok {
		return ""
	}
	return val
}

// FromLocals extracts a typed value from fiber.Ctx locals safely.
func FromLocals[T any](c fiber.Ctx, key string) (T, bool) {
	val, ok := c.Locals(key).(T)
	return val, ok
}

// WithUserID injects user ID into standard library context.Context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// FromContextUserID extracts user ID from standard library context.Context.
func FromContextUserID(ctx context.Context) (string, bool) {
	val, ok := ctx.Value(UserIDKey).(string)
	return val, ok
}

// WithValue injects a typed key-value pair into context.Context.
func WithValue[T any](ctx context.Context, key any, val T) context.Context {
	return context.WithValue(ctx, key, val)
}

// FromContext retrieves a typed value from context.Context.
func FromContext[T any](ctx context.Context, key any) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}
