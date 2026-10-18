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
	protected.Put("/me/password", h.ChangePassword)
	protected.Put("/me/pin", h.ChangePIN)

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

// ChangePassword handles self-service password updates for the authenticated user.
func (h *AuthHandler) ChangePassword(c fiber.Ctx) error {
	userID := contextutil.MustGetUserID(c)
	if userID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_UNAUTHORIZED", "pengguna tidak terautentikasi")
	}

	var req ChangePasswordRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	err := h.uc.ChangePassword(c.Context(), usecase.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		return h.mapAuthError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "password berhasil diperbarui", nil)
}

// ChangePIN handles self-service POS PIN updates or initial configuration for the authenticated user.
func (h *AuthHandler) ChangePIN(c fiber.Ctx) error {
	userID := contextutil.MustGetUserID(c)
	if userID == "" {
		return response.Fail(c, fiber.StatusUnauthorized, "AUTH_UNAUTHORIZED", "pengguna tidak terautentikasi")
	}

	var req ChangePINRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "INVALID_BODY", "format request body tidak valid")
	}

	if ok, err := response.ValidateOrFail(c, req); !ok {
		return err
	}

	err := h.uc.ChangePIN(c.Context(), usecase.ChangePINInput{
		UserID:     userID,
		CurrentPIN: req.CurrentPIN,
		NewPIN:     req.NewPIN,
	})
	if err != nil {
		return h.mapAuthError(c, err)
	}

	return response.Success(c, fiber.StatusOK, "PIN berhasil diperbarui", nil)
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
	case errors.Is(err, authdomain.ErrCurrentPasswordMismatch):
		return response.Fail(c, fiber.StatusBadRequest, "AUTH_INVALID_CURRENT_PASSWORD", "password saat ini tidak sesuai")
	case errors.Is(err, authdomain.ErrCurrentPINMismatch):
		return response.Fail(c, fiber.StatusBadRequest, "AUTH_INVALID_CURRENT_PIN", "PIN saat ini tidak sesuai")
	case errors.Is(err, authdomain.ErrCurrentPINRequired):
		return response.Fail(c, fiber.StatusBadRequest, "AUTH_CURRENT_PIN_REQUIRED", "PIN saat ini wajib diisi")
	case errors.Is(err, authdomain.ErrSamePassword):
		return response.Fail(c, fiber.StatusUnprocessableEntity, "AUTH_SAME_PASSWORD", "password baru tidak boleh sama dengan password saat ini")
	case errors.Is(err, authdomain.ErrSamePIN):
		return response.Fail(c, fiber.StatusUnprocessableEntity, "AUTH_SAME_PIN", "PIN baru tidak boleh sama dengan PIN saat ini")
	case errors.Is(err, authdomain.ErrTooManyAttempts):
		return response.Fail(c, fiber.StatusTooManyRequests, "AUTH_TOO_MANY_ATTEMPTS", "terlalu banyak percobaan login gagal, akun terkunci sementara selama 15 menit")
	case errors.Is(err, userdomain.ErrUserNotFound):
		return response.Fail(c, fiber.StatusNotFound, "USER_NOT_FOUND", "pengguna tidak ditemukan")
	default:
		return response.Fail(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "terjadi kesalahan internal pada server")
	}
}
