package money_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hanjo-erp/hanjo/pkg/money"
)

func TestMoney_Operations(t *testing.T) {
	m1 := money.IDR(50000)
	m2 := money.IDR(15000)

	// Add
	sum, err := m1.Add(m2)
	assert.NoError(t, err)
	assert.Equal(t, int64(65000), sum.Amount())
	assert.Equal(t, "Rp 65.000", sum.FormatIDR())

	// Subtract
	diff, err := m1.Subtract(m2)
	assert.NoError(t, err)
	assert.Equal(t, int64(35000), diff.Amount())

	// Negative error
	_, err = m2.Subtract(m1)
	assert.ErrorIs(t, err, money.ErrNegativeAmount)

	// Multiply
	mult := m2.Multiply(3)
	assert.Equal(t, int64(45000), mult.Amount())

	// Percentage (11% tax)
	tax := m1.Percentage(11.0)
	assert.Equal(t, int64(5500), tax.Amount())
}
