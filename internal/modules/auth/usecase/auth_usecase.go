package usecase

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	authdomain "github.com/hanjo-erp/hanjo/internal/modules/auth/domain"
	userdomain "github.com/hanjo-erp/hanjo/internal/modules/user/domain"
	"github.com/hanjo-erp/hanjo/pkg/jwt"
	"github.com/hanjo-erp/hanjo/pkg/password"
	"github.com/hanjo-erp/hanjo/pkg/ptr"
)

// AuthUsecase defines authentication & session orchestration contracts.
type AuthUsecase interface {
	Login(ctx context.Context, input LoginInput) (*AuthOutput, error)
	POSLogin(ctx context.Context, input POSLoginInput) (*AuthOutput, error)
	RefreshToken(ctx context.Context, refreshTokenString string) (*authdomain.TokenPair, error)
	Me(ctx context.Context, userID string) (*UserProfileOutput, error)
}

type authUsecase struct {
	userRepo        userdomain.UserRepository
	jwtManager      *jwt.Manager
	accessDuration  time.Duration
	refreshDuration time.Duration
	log             *zap.Logger
}

// NewAuthUsecase creates a new AuthUsecase instance.
func NewAuthUsecase(
	userRepo userdomain.UserRepository,
	jwtManager *jwt.Manager,
	accessDuration, refreshDuration time.Duration,
	log *zap.Logger,
) AuthUsecase {
	if accessDuration <= 0 {
		accessDuration = 8 * time.Hour
	}
	if refreshDuration <= 0 {
		refreshDuration = 7 * 24 * time.Hour
	}

	return &authUsecase{
		userRepo:        userRepo,
		jwtManager:      jwtManager,
		accessDuration:  accessDuration,
		refreshDuration: refreshDuration,
		log:             log,
	}
}

func (u *authUsecase) Login(ctx context.Context, input LoginInput) (*AuthOutput, error) {
	identifier := strings.TrimSpace(input.Identifier)
	if identifier == "" || input.Password == "" {
		return nil, authdomain.ErrInvalidCredentials
	}

	user, err := u.userRepo.GetByEmail(ctx, identifier)
	if err != nil {
		u.log.Warn("login failed: user not found", zap.String("identifier", identifier))
		return nil, authdomain.ErrInvalidCredentials
	}

	if err := password.Compare(user.Password, input.Password); err != nil {
		u.log.Warn("login failed: invalid password", zap.String("user_id", user.ID))
		return nil, authdomain.ErrInvalidCredentials
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return nil, authdomain.ErrAccountInactive
	}

	tokenPair, err := u.generateTokenPair(user)
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

	user, err := u.userRepo.GetByOutletAndUsername(ctx, input.OutletID, input.Username)
	if err != nil {
		u.log.Warn("pos login failed: user not found for outlet",
			zap.String("outlet_id", input.OutletID),
			zap.String("username", input.Username),
		)
		return nil, authdomain.ErrInvalidCredentials
	}

	if user.Pin == nil || *user.Pin == "" {
		return nil, authdomain.ErrInvalidPIN
	}

	if err := password.Compare(*user.Pin, input.PIN); err != nil {
		u.log.Warn("pos login failed: invalid PIN", zap.String("user_id", user.ID))
		return nil, authdomain.ErrInvalidPIN
	}

	if !strings.EqualFold(user.Status, "ACTIVE") {
		return nil, authdomain.ErrAccountInactive
	}

	tokenPair, err := u.generateTokenPair(user)
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

	return u.generateTokenPair(user)
}

func (u *authUsecase) Me(ctx context.Context, userID string) (*UserProfileOutput, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, userdomain.ErrUserNotFound
	}

	return &UserProfileOutput{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		BrandID:        user.BrandID,
		OutletID:       user.OutletID,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}, nil
}

func (u *authUsecase) generateTokenPair(user *userdomain.User) (*authdomain.TokenPair, error) {
	now := time.Now()
	accessExpiresAt := now.Add(u.accessDuration)

	accessPayload := jwt.TokenPayload{
		UserID:         user.ID,
		Email:          user.Email,
		Name:           user.Username,
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
		BrandID:        user.BrandID,
		OutletID:       user.OutletID,
		Username:       user.Username,
		Email:          user.Email,
		Status:         user.Status,
	}
}
