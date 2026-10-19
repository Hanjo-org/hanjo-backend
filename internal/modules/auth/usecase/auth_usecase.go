package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	authdomain "github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
	rbacdomain "github.com/hanjo-erp/hanjo/internal/modules/rbac/domain"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/limiter"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
)

// AuthUsecase defines authentication & session orchestration contracts.
type AuthUsecase interface {
	Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
	POSLogin(ctx context.Context, input POSLoginInput) (*AuthOutput, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*authdomain.TokenPair, error)
	Me(ctx context.Context, userID string) (*UserProfileOutput, error)
	ChangePassword(ctx context.Context, input ChangePasswordInput) error
	ChangePIN(ctx context.Context, input ChangePINInput) error
}

const (
	DefaultMaxLoginAttempts = 5
	DefaultLockoutDuration  = 15 * time.Minute
)

type authUsecase struct {
	userRepo        userdomain.UserRepository
	roleRepo        rbacdomain.RoleRepository
	jwtManager      *jwt.Manager
	rateLimiter     limiter.RateLimiter
	accessDuration  time.Duration
	refreshDuration time.Duration
	maxAttempts     int
	lockDuration    time.Duration
	log             *zap.Logger
}

// NewAuthUsecase creates a new AuthUsecase instance with brute-force protection.
func NewAuthUsecase(
	userRepo userdomain.UserRepository,
	roleRepo rbacdomain.RoleRepository,
	jwtManager *jwt.Manager,
	rateLimiter limiter.RateLimiter,
	accessDuration, refreshDuration time.Duration,
	log *zap.Logger,
) AuthUsecase {
	if accessDuration <= 0 {
		accessDuration = 8 * time.Hour
	}
	if refreshDuration <= 0 {
		refreshDuration = 7 * 24 * time.Hour
	}
	if rateLimiter == nil {
		rateLimiter = limiter.NewMemoryRateLimiter()
	}

	return &authUsecase{
		userRepo:        userRepo,
		roleRepo:        roleRepo,
		jwtManager:      jwtManager,
		rateLimiter:     rateLimiter,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		maxAttempts:     DefaultMaxLoginAttempts,
		lockDuration:    DefaultLockoutDuration,
		log:             log,
	}
}

func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	identifier := strings.TrimSpace(input.Identifier)
	if identifier == "" || input.Password == "" {
		return nil, authdomain.ErrInvalidCredentials
	}

	rateLimitKey := "auth:login:" + strings.ToLower(identifier)
	if isLocked, lockedUntil, err := u.rateLimiter.CheckLockout(ctx, rateLimitKey); err == nil && isLocked {
		u.log.Warn("login rejected: account locked due to too many failed attempts",
			zap.String("identifier", identifier),
			zap.Timep("locked_until", lockedUntil),
		)
		return nil, authdomain.ErrTooManyAttempts
	}

	user, err := u.userRepo.GetByEmail(ctx, identifier)
	if err != nil {
		u.log.Warn("login failed: user not found", zap.String("identifier", identifier))
		if _, _, lockedUntil, _ := u.rateLimiter.RecordFailedAttempt(ctx, rateLimitKey, u.maxAttempts, u.lockDuration); lockedUntil != nil {
			u.log.Warn("login failed and account locked out", zap.String("identifier", identifier), zap.Timep("locked_until", lockedUntil))
			return nil, authdomain.ErrTooManyAttempts
		}
		return nil, authdomain.ErrInvalidCredentials
	}

	if err := password.Compare(user.Password, input.Password); err != nil {
		u.log.Warn("login failed: invalid password", zap.String("user_id", user.ID))
		if _, _, lockedUntil, _ := u.rateLimiter.RecordFailedAttempt(ctx, rateLimitKey, u.maxAttempts, u.lockDuration); lockedUntil != nil {
			u.log.Warn("login failed: invalid password and account locked out", zap.String("identifier", identifier), zap.Timep("locked_until", lockedUntil))
			return nil, authdomain.ErrTooManyAttempts
		}
		return nil, authdomain.ErrInvalidCredentials
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return nil, authdomain.ErrAccountInactive
	}

	// Reset failed attempts on successful login
	_ = u.rateLimiter.Reset(ctx, rateLimitKey)

	tokenPair, err := u.generateTokenPair(ctx, user)
	if err != nil {
		u.log.Error("failed to generate token pair", zap.Error(err), zap.String("user_id", user.ID))
		return nil, err
	}

	return &AuthOutput{
		User:   toUserOutput(user),
		Tokens: *tokenPair,
	}, nil
}

func (u *authUsecase) POSLogin(ctx context.Context, input POSLoginInput) (*AuthOutput, error) {
	if input.OutletID == "" {
		return nil, authdomain.ErrOutletRequired
	}
	if input.Username == "" || input.PIN == "" {
		return nil, authdomain.ErrInvalidCredentials
	}

	rateLimitKey := fmt.Sprintf("auth:pos:%s:%s", input.OutletID, strings.ToLower(strings.TrimSpace(input.Username)))
	if isLocked, lockedUntil, err := u.rateLimiter.CheckLockout(ctx, rateLimitKey); err == nil && isLocked {
		u.log.Warn("pos login rejected: cashier PIN locked due to too many failed attempts",
			zap.String("outlet_id", input.OutletID),
			zap.String("username", input.Username),
			zap.Timep("locked_until", lockedUntil),
		)
		return nil, authdomain.ErrTooManyAttempts
	}

	user, err := u.userRepo.GetByOutletAndUsername(ctx, input.OutletID, input.Username)
	if err != nil {
		u.log.Warn("pos login failed: user not found for outlet",
			zap.String("outlet_id", input.OutletID),
			zap.String("username", input.Username),
		)
		if _, _, lockedUntil, _ := u.rateLimiter.RecordFailedAttempt(ctx, rateLimitKey, u.maxAttempts, u.lockDuration); lockedUntil != nil {
			u.log.Warn("pos login failed and cashier locked out", zap.String("outlet_id", input.OutletID), zap.String("username", input.Username), zap.Timep("locked_until", lockedUntil))
			return nil, authdomain.ErrTooManyAttempts
		}
		return nil, authdomain.ErrInvalidCredentials
	}

	if user.Pin == nil || *user.Pin == "" {
		if _, _, lockedUntil, _ := u.rateLimiter.RecordFailedAttempt(ctx, rateLimitKey, u.maxAttempts, u.lockDuration); lockedUntil != nil {
			return nil, authdomain.ErrTooManyAttempts
		}
		return nil, authdomain.ErrInvalidPIN
	}

	if err := password.Compare(*user.Pin, input.PIN); err != nil {
		u.log.Warn("pos login failed: invalid PIN", zap.String("user_id", user.ID))
		if _, _, lockedUntil, _ := u.rateLimiter.RecordFailedAttempt(ctx, rateLimitKey, u.maxAttempts, u.lockDuration); lockedUntil != nil {
			u.log.Warn("pos login failed: invalid PIN and cashier locked out", zap.String("outlet_id", input.OutletID), zap.String("username", input.Username), zap.Timep("locked_until", lockedUntil))
			return nil, authdomain.ErrTooManyAttempts
		}
		return nil, authdomain.ErrInvalidPIN
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return nil, authdomain.ErrAccountInactive
	}

	// Reset failed attempts on successful login
	_ = u.rateLimiter.Reset(ctx, rateLimitKey)

	tokenPair, err := u.generateTokenPair(ctx, user)
	if err != nil {
		u.log.Error("failed to generate token pair for pos", zap.Error(err), zap.String("user_id", user.ID))
		return nil, err
	}

	return &AuthOutput{
		User:   toUserOutput(user),
		Tokens: *tokenPair,
	}, nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, refreshTokenString string) (*authdomain.TokenPair, error) {
	if refreshTokenString == "" {
		return nil, authdomain.ErrInvalidToken
	}

	claims, err := u.jwtManager.ParseToken(refreshTokenString)
	if err != nil {
		return nil, authdomain.ErrTokenExpired
	}

	user, err := u.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, authdomain.ErrInvalidToken
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return nil, authdomain.ErrAccountInactive
	}

	return u.generateTokenPair(ctx, user)
}

func (u *authUsecase) Me(ctx context.Context, userID string) (*UserProfileOutput, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, userdomain.ErrUserNotFound
	}

	return &UserProfileOutput{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		RoleID:         user.RoleID,
		BrandID:        user.BrandID,
		OutletID:       user.OutletID,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}, nil
}

func (u *authUsecase) generateTokenPair(ctx context.Context, user *userdomain.User) (*authdomain.TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(u.accessDuration)

	var roleCode string
	var roleID string
	if user.RoleID != nil && *user.RoleID != "" {
		roleID = *user.RoleID
		if u.roleRepo != nil {
			if r, err := u.roleRepo.GetByID(ctx, roleID); err == nil && r != nil {
				roleCode = r.Code
			}
		}
	}

	accessPayload := jwt.TokenPayload{
		UserID:         user.ID,
		Email:          user.Email,
		Name:           user.Username,
		Role:           roleCode,
		RoleID:         roleID,
		OrganizationID: user.OrganizationID,
		BrandID:        ptr.Val(user.BrandID),
		OutletID:       ptr.Val(user.OutletID),
		Duration:       u.accessDuration,
	}

	accessToken, err := u.jwtManager.GenerateToken(accessPayload)
	if err != nil {
		return nil, err
	}

	refreshPayload := jwt.TokenPayload{
		UserID:         user.ID,
		Email:          user.Email,
		Role:           roleCode,
		RoleID:         roleID,
		OrganizationID: user.OrganizationID,
		BrandID:        ptr.Val(user.BrandID),
		OutletID:       ptr.Val(user.OutletID),
		Duration:       u.refreshDuration,
	}

	refreshToken, err := u.jwtManager.GenerateToken(refreshPayload)
	if err != nil {
		return nil, err
	}

	return &authdomain.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.accessDuration.Seconds()),
		ExpiresAt:    accessExpiresAt,
	}, nil
}

func toUserOutput(user *userdomain.User) UserOutput {
	return UserOutput{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		RoleID:         user.RoleID,
		BrandID:        user.BrandID,
		OutletID:       user.OutletID,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
	}
}

// ChangePassword updates the authenticated user's password after verifying their current password.
func (u *authUsecase) ChangePassword(ctx context.Context, input ChangePasswordInput) error {
	user, err := u.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return userdomain.ErrUserNotFound
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return authdomain.ErrAccountInactive
	}

	// Verify current password
	if err := password.Compare(user.Password, input.CurrentPassword); err != nil {
		return authdomain.ErrCurrentPasswordMismatch
	}

	// Ensure new password is not identical to current password
	if err := password.Compare(user.Password, input.NewPassword); err == nil {
		return authdomain.ErrSamePassword
	}

	hashedPassword, err := password.Hash(input.NewPassword)
	if err != nil {
		u.log.Error("failed to hash new password", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	user.Password = hashedPassword
	user.UpdatedAt = time.Now().UTC()

	if err := u.userRepo.Update(ctx, user); err != nil {
		u.log.Error("failed to update user password in repository", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	_ = u.rateLimiter.Reset(ctx, "auth:login:"+strings.ToLower(user.Email))
	_ = u.rateLimiter.Reset(ctx, "auth:login:"+strings.ToLower(user.Username))

	u.log.Info("user password updated successfully", zap.String("user_id", user.ID))
	return nil
}

// ChangePIN updates or sets the authenticated user's POS PIN after verifying their current PIN if one exists.
func (u *authUsecase) ChangePIN(ctx context.Context, input ChangePINInput) error {
	user, err := u.userRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return userdomain.ErrUserNotFound
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return authdomain.ErrAccountInactive
	}

	// If user already has a PIN configured, verify current PIN
	if user.Pin != nil && *user.Pin != "" {
		if input.CurrentPIN == nil || *input.CurrentPIN == "" {
			return authdomain.ErrCurrentPINRequired
		}

		if err := password.Compare(*user.Pin, *input.CurrentPIN); err != nil {
			return authdomain.ErrCurrentPINMismatch
		}

		// Ensure new PIN is not identical to current PIN
		if err := password.Compare(*user.Pin, input.NewPIN); err == nil {
			return authdomain.ErrSamePIN
		}
	}

	hashedPIN, err := password.Hash(input.NewPIN)
	if err != nil {
		u.log.Error("failed to hash new PIN", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	user.Pin = &hashedPIN
	user.UpdatedAt = time.Now().UTC()

	if err := u.userRepo.Update(ctx, user); err != nil {
		u.log.Error("failed to update user PIN in repository", zap.Error(err), zap.String("user_id", user.ID))
		return err
	}

	if user.OutletID != nil && *user.OutletID != "" {
		_ = u.rateLimiter.Reset(ctx, fmt.Sprintf("auth:pos:%s:%s", *user.OutletID, strings.ToLower(user.Username)))
	}

	u.log.Info("user PIN updated successfully", zap.String("user_id", user.ID))
	return nil
}
