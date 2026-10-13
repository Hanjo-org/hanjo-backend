package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/hanjo-erp/hanjo/pkg/response"
)

type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	UserEmailKey      contextKey = "user_email"
	OrganizationIDKey contextKey = "organization_id"
	OutletIDKey       contextKey = "outlet_id"
	UserRoleKey       contextKey = "user_role"
)

// Claims represents the JWT claims for Hanjo.
type Claims struct {
	jwt.RegisteredClaims
	UserID         string `json:"user_id"`
	Email          string `json:"email"`
	OrganizationID string `json:"organization_id"`
	OutletID       string `json:"outlet_id,omitempty"`
	Role           string `json:"role"`
}

// Auth returns a middleware that validates JWT tokens and injects user context.
func Auth(jwtSecret string, log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Unauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				response.Unauthorized(w, "invalid authorization header format")
				return
			}

			tokenString := parts[1]

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !token.Valid {
				log.Debug().Err(err).Msg("invalid JWT token")
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			// Inject user info into context
			ctx := r.Context()
			if userID, err := uuid.Parse(claims.UserID); err == nil {
				ctx = context.WithValue(ctx, UserIDKey, userID)
			}
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			if orgID, err := uuid.Parse(claims.OrganizationID); err == nil {
				ctx = context.WithValue(ctx, OrganizationIDKey, orgID)
			}
			if claims.OutletID != "" {
				if outletID, err := uuid.Parse(claims.OutletID); err == nil {
					ctx = context.WithValue(ctx, OutletIDKey, outletID)
				}
			}
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok
}

// GetOrganizationID extracts the organization ID from the request context.
func GetOrganizationID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(OrganizationIDKey).(uuid.UUID)
	return id, ok
}

// GetOutletID extracts the outlet ID from the request context.
func GetOutletID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(OutletIDKey).(uuid.UUID)
	return id, ok
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(UserRoleKey).(string)
	return role
}

// RequireRole returns a middleware that checks if the user has one of the required roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, r := range roles {
		roleSet[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r.Context())
			if !roleSet[userRole] {
				response.Forbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
