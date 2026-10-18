package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordMismatch = errors.New("password does not match")
	ErrPasswordTooLong  = errors.New("password exceeds maximum allowed length of 72 bytes")
)

const (
	// DefaultCost is the bcrypt hashing cost factor (10 is standard).
	DefaultCost = bcrypt.DefaultCost
)

// Hash hashes plaintext using bcrypt with DefaultCost.
func Hash(plain string) (string, error) {
	return HashWithCost(plain, DefaultCost)
}

// HashWithCost hashes plaintext using bcrypt with a specified cost factor.
func HashWithCost(plain string, cost int) (string, error) {
	if len(plain) > 72 {
		return "", ErrPasswordTooLong
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare verifies whether a plaintext matches a bcrypt hashed password.
func Compare(hashedPassword, plain string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return err
	}
	return nil
}
