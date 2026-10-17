package password_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/hanjo-erp/hanjo/pkg/password"
)

func TestBcrypt_HashAndCompare(t *testing.T) {
	plain := "SecretPass123!"
	hashed, err := password.Hash(plain)
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)

	// Verify valid password
	err = password.Compare(hashed, plain)
	assert.NoError(t, err)

	// Verify wrong password
	err = password.Compare(hashed, "WrongPassword")
	assert.ErrorIs(t, err, password.ErrPasswordMismatch)
}

func TestBcrypt_PINHashAndCompare(t *testing.T) {
	pin := "123456"
	hashed, err := password.Hash(pin)
	require.NoError(t, err)
	assert.NotEmpty(t, hashed)

	err = password.Compare(hashed, pin)
	assert.NoError(t, err)

	err = password.Compare(hashed, "654321")
	assert.ErrorIs(t, err, password.ErrPasswordMismatch)
}

func TestBcrypt_PasswordTooLong(t *testing.T) {
	tooLong := strings.Repeat("A", 73)
	_, err := password.Hash(tooLong)
	assert.ErrorIs(t, err, password.ErrPasswordTooLong)
}

func TestBcrypt_InvalidHashFormat(t *testing.T) {
	err := password.Compare("invalid-hash-string", "123456")
	assert.Error(t, err)
}
