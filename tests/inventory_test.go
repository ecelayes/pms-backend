package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ecelayes/pms-backend/internal/entity"
	"github.com/ecelayes/pms-backend/internal/service"
)

func TestInventoryLogic(t *testing.T) {
	inventoryService := service.NewInventoryService()

	jan1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	jan10 := time.Date(2025, 1, 10, 0, 0, 0, 0, time.UTC)
	jan15 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	jan20 := time.Date(2025, 1, 20, 0, 0, 0, 0, time.UTC)
	jan31 := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	feb1 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)

	baseRule := entity.PriceRule{
		Start: jan1,
		End:   jan31,
		Price: 100.0,
	}

	t.Run("Case 1: No Overlap (Independent)", func(t *testing.T) {
		feb1 := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
		feb5 := time.Date(2025, 2, 5, 0, 0, 0, 0, time.UTC)
		newRule := entity.PriceRule{Start: feb1, End: feb5, Price: 200.0}

		result := inventoryService.ResolveRuleConflicts([]entity.PriceRule{baseRule}, newRule)

		assert.Len(t, result, 2)
	})

	t.Run("Case 2: Internal Split (Middle Cut)", func(t *testing.T) {
		newRule := entity.PriceRule{Start: jan10, End: jan15, Price: 200.0}

		result := inventoryService.ResolveRuleConflicts([]entity.PriceRule{baseRule}, newRule)

		assert.Len(t, result, 3)
		
		assert.Equal(t, 200.0, result[0].Price, "First element should be the new rule")
		
		var left, right entity.PriceRule
		foundLeft, foundRight := false, false
		
		for _, r := range result {
			if r.Start.Equal(jan1) && r.End.Equal(jan10) {
				left = r
				foundLeft = true
			}
			if r.Start.Equal(jan15) && r.End.Equal(jan31) {
				right = r
				foundRight = true
			}
		}

		assert.True(t, foundLeft, "Left fragment [Jan 1 - Jan 10] missing")
		assert.Equal(t, 100.0, left.Price)

		assert.True(t, foundRight, "Right fragment [Jan 15 - Jan 31] missing")
		assert.Equal(t, 100.0, right.Price)
	})

	t.Run("Case 3: Complete Overwrite", func(t *testing.T) {
		newRule := entity.PriceRule{Start: jan1, End: jan31, Price: 500.0}

		result := inventoryService.ResolveRuleConflicts([]entity.PriceRule{baseRule}, newRule)

		assert.Len(t, result, 1)
		assert.Equal(t, 500.0, result[0].Price)
	})

	t.Run("Case 4: Partial Overlap Left", func(t *testing.T) {
		dec31 := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		newRule := entity.PriceRule{Start: dec31, End: jan10, Price: 300.0}

		result := inventoryService.ResolveRuleConflicts([]entity.PriceRule{baseRule}, newRule)

		assert.Len(t, result, 2)
		
		foundRight := false
		for _, r := range result {
			if r.Start.Equal(jan10) && r.End.Equal(jan31) {
				foundRight = true
				break
			}
		}
		assert.True(t, foundRight, "Should create right fragment [10-31]")
	})

	t.Run("Case 5: Partial Overlap Right", func(t *testing.T) {
		newRule := entity.PriceRule{Start: jan20, End: feb1, Price: 300.0}

		result := inventoryService.ResolveRuleConflicts([]entity.PriceRule{baseRule}, newRule)

		assert.Len(t, result, 2)
		
		foundLeft := false
		for _, r := range result {
			if r.Start.Equal(jan1) && r.End.Equal(jan20) {
				foundLeft = true
				break
			}
		}
		assert.True(t, foundLeft, "Should create left fragment [1-20]")
	})
}
