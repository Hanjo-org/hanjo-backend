package jwt_test

import (
	"testing"
	"time"

	"github.com/hanjo-erp/hanjo/pkg/jwt"
)

func TestRS256_GenerateAndParseToken(t *testing.T) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate RSA keypair: %v", err)
	}

	mgr := jwt.NewRSAManager(privKey, pubKey, "https://api.hanjo-erp.com", "https://api.hanjo-erp.com")

	payload := jwt.TokenPayload{
		UserID:   "user-123",
		NIK:      "10928374",
		Name:     "Budi Santoso",
		Role:     "FIELD_OFFICER",
		UnitID:   "unit-abc",
		ShiftID:  "shift-xyz",
		Duration: 1 * time.Hour,
	}

	tokenStr, err := mgr.GenerateToken(payload)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	claims, err := mgr.ParseToken(tokenStr)
	if err != nil {
		t.Fatalf("failed to parse valid token: %v", err)
	}

	if claims.Subject != payload.UserID || claims.UserID != payload.UserID {
		t.Errorf("expected subject %s, got %s", payload.UserID, claims.Subject)
	}
	if claims.NIK != payload.NIK {
		t.Errorf("expected NIK %s, got %s", payload.NIK, claims.NIK)
	}
	if claims.Role != payload.Role {
		t.Errorf("expected Role %s, got %s", payload.Role, claims.Role)
	}
	if claims.UnitID != payload.UnitID {
		t.Errorf("expected UnitID %s, got %s", payload.UnitID, claims.UnitID)
	}
	if claims.ShiftID != payload.ShiftID {
		t.Errorf("expected ShiftID %s, got %s", payload.ShiftID, claims.ShiftID)
	}
}

func TestRS256_ExpiredToken(t *testing.T) {
	privKey, pubKey, err := jwt.GenerateRSAKeyPair(2048)
	if err != nil {
		t.Fatalf("failed to generate RSA keypair: %v", err)
	}

	mgr := jwt.NewRSAManager(privKey, pubKey, "iss", "aud")

	payload := jwt.TokenPayload{
		UserID:   "user-123",
		Duration: -1 * time.Second, // already expired
	}

	tokenStr, err := mgr.GenerateToken(payload)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = mgr.ParseToken(tokenStr)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}
