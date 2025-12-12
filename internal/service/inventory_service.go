package service

import (
	"github.com/ecelayes/pms-backend/internal/entity"
)

type InventoryService struct{}

func NewInventoryService() *InventoryService {
	return &InventoryService{}
}

func (s *InventoryService) ResolveRuleConflicts(existingRules []entity.PriceRule, newRule entity.PriceRule) []entity.PriceRule {
	var result []entity.PriceRule

	result = append(result, newRule)

	for _, existing := range existingRules {
		isBefore := existing.End.Before(newRule.Start) || existing.End.Equal(newRule.Start)
		isAfter := existing.Start.After(newRule.End) || existing.Start.Equal(newRule.End)

		if isBefore || isAfter {
			result = append(result, existing)
			continue
		}

		if existing.Start.Before(newRule.Start) {
			leftFragment := entity.PriceRule{
				BaseEntity: existing.BaseEntity,
				RoomTypeID: existing.RoomTypeID,
				Start:      existing.Start,
				End:        newRule.Start,
				Price:      existing.Price,
			}
			result = append(result, leftFragment)
		}

		if existing.End.After(newRule.End) {
			rightFragment := entity.PriceRule{
				BaseEntity: existing.BaseEntity,
				RoomTypeID: existing.RoomTypeID,
				Start:      newRule.End,
				End:        existing.End,
				Price:      existing.Price,
			}
			result = append(result, rightFragment)
		}
	}

	return result
}
