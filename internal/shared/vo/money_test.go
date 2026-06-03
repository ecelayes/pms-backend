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
