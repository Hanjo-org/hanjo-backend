package delivery

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/hanjo-erp/hanjo/internal/middleware"
	authdomain "github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
	"github.com/hanjo-erp/hanjo/internal/modules/auth/usecase"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/contextutil"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/response"
)

type AuthHandler struct {
	uc           usecase.AuthUsecase
	cookieName   string
	secureCookie bool
}

// RegisterAuthRoutes registers all public and protected authentication routes.
func RegisterAuthRoutes(
	router fiber.Router,
	uc usecase.AuthUsecase,
	jwtManager *jwt.Manager,
	cookieName string,
	secureCookie bool,
) *AuthHandler {
	if cookieName == "" {
		cookieName = "refresh_token"
	}

	h := &AuthHandler{
		uc:           uc,
		cookieName:   cookieName,
		secureCookie: secureCookie,
	}

	authGroup := router.Group("/auth")
	authGroup.Post("/login", h.Login)
	authGroup.Post("/pos/login", h.POSLogin)
	authGroup.Post("/refresh", h.RefreshToken)

	// Protected routes
	protected := authGroup.Group("", middleware.Auth(jwtManager))
	protected.Post("/logout", h.Logout)
	protected.Get("/me", h.Me)

	return h
}

// Login handles standard email/username + password login.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	res, err := h.uc.Login(c.Context(), usecase.LoginInput{
		Identifier: req.Identifier,
		Password:   req.Password,
	})
	if err != nil {
		return h.mapAuthError(c, err)
	}

	h.setRefreshTokenCookie(c, res.Tokens.RefreshToken)
	return response.Success(c, fiber.StatusOK, "login berhasil", res)
}

// POSLogin handles fast PIN login for cashiers/baristas at a physical outlet.
func (h *AuthHandler) POSLogin(c fiber.Ctx) error {
	var req POSLoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	res, err := h.uc.POSLogin(c.Context(), usecase.POSLoginInput{
		OutletID: req.OutletID,
		Username: req.Username,
		PIN:      req.PIN,
	})
	if err != nil {
		return h.mapAuthError(c, err)
	}

	h.setRefreshTokenCookie(c, res.Tokens.RefreshToken)
	return response.Success(c, fiber.StatusOK, "pos login berhasil", res)
}

// RefreshToken rotates and issues a new access token using the refresh cookie or payload.
func (h *AuthHandler) RefreshToken(c fiber.Ctx) error {
	refreshToken := c.Cookies(h.cookieName)
	if refreshToken == "" {
		var req RefreshTokenRequest
		_ = c.Bind().Body(&req)
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_REFRESH_TOKEN_REQUIRED", "refresh token tidak ditemukan")
	}

	tokens, err := h.uc.RefreshToken(c.Context(), refreshToken)
	if err != nil {
		return h.mapAuthError(c, err)
	}

	h.setRefreshTokenCookie(c, tokens.RefreshToken)
	return response.Success(c, fiber.StatusOK, "token berhasil diperbarui", tokens)
}

// Logout clears the refresh token cookie and ends the current session.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	c.ClearCookie(h.cookieName)
	return response.Success(c, fiber.StatusOK, "logout berhasil", nil)
}

// Me retrieves the authenticated user profile and organizational context.
func (h *AuthHandler) Me(c fiber.Ctx) error {
	userID := contextutil.MustGetUserID(c)
	if userID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_UNAUTHORIZED", "pengguna tidak terautentikasi")
	}

	profile, err := h.uc.Me(c.Context(), userID)
	if err != nil {
		if errors.Is(err, userdomain.ErrUserNotFound) {
			return response.Fail(c, fiber.StatusNotFound, "USER_NOT_FOUND", "pengguna tidak ditemukan")
		}
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "terjadi kesalahan internal server")
	}

	return response.Success(c, fiber.StatusOK, "profil berhasil diambil", profile)
}

func (h *AuthHandler) setRefreshTokenCookie(c fiber.Ctx, token string) {
	if token == "" {
		return
	}
	c.Cookie(&fiber.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   h.secureCookie,
		SameSite: "Lax",
		Path:     "/",
	})
}

func (h *AuthHandler) mapAuthError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, authdomain.ErrInvalidCredentials):
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_INVALID_CREDENTIALS", "kredensial login tidak valid")
	case errors.Is(err, authdomain.ErrInvalidPIN):
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_INVALID_PIN", "PIN kasir tidak valid")
	case errors.Is(err, authdomain.ErrAccountInactive):
		return response.Fail(c, fiber.StatusForbidden, "AUTH_ACCOUNT_INACTIVE", "akun pengguna dinonaktifkan")
	case errors.Is(err, authdomain.ErrOutletRequired):
		return response.Fail(c, fiber.StatusBadRequest, "AUTH_OUTLET_REQUIRED", "outlet_id diperlukan untuk pos login")
	case errors.Is(err, authdomain.ErrTokenExpired):
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_TOKEN_EXPIRED", "sesi login telah kedaluwarsa")
	case errors.Is(err, authdomain.ErrInvalidToken):
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_INVALID_TOKEN", "token autentikasi tidak valid")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "terjadi kesalahan internal pada server")
	}
}
