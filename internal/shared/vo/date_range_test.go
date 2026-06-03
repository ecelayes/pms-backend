package vo_test

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/stretchr/testify/assert"
)

func TestDateRange(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	dayAfter := now.Add(48 * time.Hour)

	t.Run("Create", func(t *testing.T) {
		dr, err := vo.NewDateRange(now, tomorrow)
		assert.NoError(t, err)
		assert.Equal(t, now, dr.Start())
		assert.Equal(t, tomorrow, dr.End())

		_, err = vo.NewDateRange(tomorrow, now)
		assert.Error(t, err)
	})

	t.Run("DaysCount", func(t *testing.T) {
		dr, _ := vo.NewDateRange(now, dayAfter)
		assert.Equal(t, 2, dr.DaysCount())
	})

	t.Run("Overlaps", func(t *testing.T) {
		// Range A: [Today, Tomorrow]
		drA, _ := vo.NewDateRange(now, tomorrow)
		
		// Range B: [Tomorrow, DayAfter]
		drB, _ := vo.NewDateRange(tomorrow, dayAfter)

		// Range C: [Today-1hr, Today+1hr] (Overlaps A)
		drC, _ := vo.NewDateRange(now.Add(-time.Hour), now.Add(time.Hour))

		assert.False(t, drA.Overlaps(drB), "Should not overlap if end == start (usually check-out day)")
		assert.True(t, drA.Overlaps(drC), "Should overlap")
	})
}
