package password_test

import (
	"testing"

	"github.com/hanjo-erp/hanjo/pkg/password"
)

func TestArgon2id_HashAndCompare(t *testing.T) {
	// Use lighter params in unit tests for speed while preserving logic
	testParams := &password.Argon2idParams{
		Memory:      8 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	plain := "SecretPass123!"
	hashed, err := password.HashWithParams(plain, testParams)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if err := password.Compare(hashed, plain); err != nil {
		t.Errorf("expected matching password, got error: %v", err)
	}

	if err := password.Compare(hashed, "WrongPassword"); err == nil {
		t.Errorf("expected error for wrong password, got nil")
	}
}

func TestArgon2id_PINHashAndCompare(t *testing.T) {
	pin := "123456"
	testParams := &password.Argon2idParams{
		Memory:      4 * 1024,
		Iterations:  1,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hashed, err := password.HashWithParams(pin, testParams)
	if err != nil {
		t.Fatalf("unexpected error hashing PIN: %v", err)
	}

	if err := password.Compare(hashed, pin); err != nil {
		t.Errorf("expected matching PIN, got error: %v", err)
	}

	if err := password.Compare(hashed, "654321"); err == nil {
		t.Errorf("expected error for wrong PIN, got nil")
	}
}

func TestArgon2id_InvalidHashFormat(t *testing.T) {
	err := password.Compare("invalid-hash-string", "123456")
	if err == nil {
		t.Errorf("expected error for invalid hash format, got nil")
	}
}
