package domain

import (
	"errors"
	"fmt"
	"math"
)

// Money is a value object representing monetary amounts.
// It stores amounts in the smallest currency unit (e.g., cents/rupiah)
// to avoid floating-point precision issues.
type Money struct {
	amount   int64  // Amount in smallest unit (e.g., 1 Rupiah = 1)
	currency string // ISO 4217 currency code
}

var (
	ErrNegativeAmount     = errors.New("money: amount cannot be negative")
	ErrCurrencyMismatch   = errors.New("money: currency mismatch")
	ErrDivisionByZero     = errors.New("money: division by zero")
)

// NewMoney creates a new Money value object.
// Amount is in the smallest currency unit (e.g., Rupiah, not thousands).
func NewMoney(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if currency == "" {
		currency = "IDR"
	}
	return Money{amount: amount, currency: currency}, nil
}

// MustNewMoney creates a new Money and panics on error (for testing/seeding).
func MustNewMoney(amount int64, currency string) Money {
	m, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}
	return m
}

// IDR is a convenience constructor for Indonesian Rupiah.
func IDR(amount int64) Money {
	return MustNewMoney(amount, "IDR")
}

// Zero returns a zero-value Money in the given currency.
func Zero(currency string) Money {
	return Money{amount: 0, currency: currency}
}

// Amount returns the raw amount in smallest currency unit.
func (m Money) Amount() int64 {
	return m.amount
}

// Currency returns the ISO 4217 currency code.
func (m Money) Currency() string {
	return m.currency
}

// IsZero returns true if the amount is zero.
func (m Money) IsZero() bool {
	return m.amount == 0
}

// Add adds two Money values. They must share the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract subtracts another Money from this one.
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}
	result := m.amount - other.amount
	if result < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{amount: result, currency: m.currency}, nil
}

// Multiply multiplies the amount by a factor (e.g., quantity).
func (m Money) Multiply(factor int) Money {
	return Money{amount: m.amount * int64(factor), currency: m.currency}
}

// MultiplyDecimal multiplies by a decimal factor (e.g., tax rate 0.10).
// Uses rounding to nearest integer.
func (m Money) MultiplyDecimal(factor float64) Money {
	result := math.Round(float64(m.amount) * factor)
	return Money{amount: int64(result), currency: m.currency}
}

// Percentage calculates a percentage of the money amount.
// e.g., m.Percentage(10) returns 10% of the amount.
func (m Money) Percentage(percent float64) Money {
	return m.MultiplyDecimal(percent / 100.0)
}

// GreaterThan returns true if this Money is greater than the other.
func (m Money) GreaterThan(other Money) bool {
	return m.currency == other.currency && m.amount > other.amount
}

// Equals returns true if both amount and currency match.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// String returns a human-readable representation (e.g., "IDR 25000").
func (m Money) String() string {
	return fmt.Sprintf("%s %d", m.currency, m.amount)
}

// FormatIDR formats the amount as Indonesian Rupiah display string.
// e.g., 25000 → "Rp 25.000"
func (m Money) FormatIDR() string {
	if m.currency != "IDR" {
		return m.String()
	}

	str := fmt.Sprintf("%d", m.amount)
	n := len(str)
	if n <= 3 {
		return fmt.Sprintf("Rp %s", str)
	}

	var result []byte
	for i, ch := range str {
		if i > 0 && (n-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(ch))
	}
	return fmt.Sprintf("Rp %s", string(result))
}
