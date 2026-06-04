package vo_test

import (
	"testing"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/stretchr/testify/assert"
)

func TestMoney(t *testing.T) {
	m1 := vo.NewMoney(1000, "USD")
	m2 := vo.NewMoney(500, "USD")
	mDiff := vo.NewMoney(500, "EUR")

	t.Run("Add", func(t *testing.T) {
		sum, err := m1.Add(m2)
		assert.NoError(t, err)
		assert.Equal(t, int64(1500), sum.Amount())
		assert.Equal(t, "USD", sum.Currency())

		_, err = m1.Add(mDiff)
		assert.ErrorIs(t, err, vo.ErrMismatchingCurrency)
	})

	t.Run("Sub", func(t *testing.T) {
		diff, err := m1.Sub(m2)
		assert.NoError(t, err)
		assert.Equal(t, int64(500), diff.Amount())

		_, err = m1.Sub(mDiff)
		assert.ErrorIs(t, err, vo.ErrMismatchingCurrency)
	})

	t.Run("Multiply", func(t *testing.T) {
		prod := m1.Multiply(2)
		assert.Equal(t, int64(2000), prod.Amount())
	})
}

func TestMoneyString(t *testing.T) {
	m := vo.NewMoney(1500, "USD")
	s := m.String()
	
	if s != "1500 USD" {
		t.Errorf("Expected '1500 USD', got '%s'", s)
	}
}

func TestMoneyStringDifferentCurrency(t *testing.T) {
	testCases := []struct {
		amount   int64
		currency string
		expected string
	}{
		{1000, "USD", "1000 USD"},
		{2500, "EUR", "2500 EUR"},
		{0, "GBP", "0 GBP"},
		{99999, "JPY", "99999 JPY"},
	}
	
	for _, tc := range testCases {
		m := vo.NewMoney(tc.amount, tc.currency)
		s := m.String()
		if s != tc.expected {
			t.Errorf("For %d %s: expected '%s', got '%s'", tc.amount, tc.currency, tc.expected, s)
		}
	}
}
