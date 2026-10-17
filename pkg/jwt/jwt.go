package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrInvalidKey     = errors.New("invalid RSA key")
	ErrTokenExpired   = errors.New("token has expired")
	ErrInvalidMethod  = errors.New("unexpected signing method")
)

// Claims represents the JWT payload.
type Claims struct {
	UserID         string `json:"user_id,omitempty"` // backward compat alias for sub
	Email          string `json:"email,omitempty"`
	Name           string `json:"name,omitempty"`
	Role           string `json:"role,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	BrandID        string `json:"brand_id,omitempty"`
	OutletID       string `json:"outlet_id,omitempty"`
	NIK            string `json:"nik,omitempty"`
	UnitID         string `json:"unit_id,omitempty"`
	ShiftID        string `json:"shift_id,omitempty"`
	jwt.RegisteredClaims
}

// TokenPayload carries claims to be embedded into newly issued JWT tokens.
type TokenPayload struct {
	UserID         string
	Email          string
	Name           string
	Role           string
	OrganizationID string
	BrandID        string
	OutletID       string
	NIK            string
	UnitID         string
	ShiftID        string
	Duration       time.Duration
}

// Manager handles RSA RS256 token creation and validation.
type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
	audience   string
}

// NewRSAManager creates a new RS256 JWT Manager.
func NewRSAManager(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, issuer, audience string) *Manager {
	return &Manager{
		privateKey: privateKey,
		publicKey:  publicKey,
		issuer:     issuer,
		audience:   audience,
	}
}

// GenerateToken issues a signed RS256 JWT token based on the provided payload and duration.
func (m *Manager) GenerateToken(p TokenPayload) (string, error) {
	if m.privateKey == nil {
		return "", errors.New("rsa private key is required to sign tokens")
	}

	now := time.Now()
	claims := Claims{
		UserID:         p.UserID,
		Email:          p.Email,
		Name:           p.Name,
		Role:           p.Role,
		OrganizationID: p.OrganizationID,
		BrandID:        p.BrandID,
		OutletID:       p.OutletID,
		NIK:            p.NIK,
		UnitID:         p.UnitID,
		ShiftID:        p.ShiftID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Audience:  jwt.ClaimStrings{m.audience},
			Subject:   p.UserID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.Duration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

// ParseToken validates and extracts claims from an RS256 token string.
func (m *Manager) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("%w: %v", ErrInvalidMethod, t.Header["alg"])
		}
		return m.publicKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Ensure UserID field is populated from Subject if empty
	if claims.UserID == "" {
		claims.UserID = claims.Subject
	}

	return claims, nil
}

// GenerateRSAKeyPair creates a new 2048-bit RSA keypair (useful for local development & testing).
func GenerateRSAKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

// ParsePrivateKeyFromPEM parses an RSA private key from PEM bytes (PKCS#1 or PKCS#8).
func ParsePrivateKeyFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidKey
	}

	// Try PKCS#1
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS#8
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := parsedKey.(*rsa.PrivateKey); ok {
			return rsaKey, nil
		}
	}

	return nil, ErrInvalidKey
}

// ParsePublicKeyFromPEM parses an RSA public key from PEM bytes (PKIX or PKCS#1).
func ParsePublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, ErrInvalidKey
	}

	// Try PKIX
	parsedKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err == nil {
		if rsaKey, ok := parsedKey.(*rsa.PublicKey); ok {
			return rsaKey, nil
		}
	}

	// Try PKCS#1
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, ErrInvalidKey
}
