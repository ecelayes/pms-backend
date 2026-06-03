package domain

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type RatePlanRepository interface {
	Save(ctx context.Context, rp *RatePlan) error
	FindByID(ctx context.Context, id string) (*RatePlan, error)
	FindByPropertyID(ctx context.Context, propertyID string) ([]*RatePlan, error)
	Delete(ctx context.Context, id string) error
}
type PriceRuleRepository interface {
	Save(ctx context.Context, pr *PriceRule) error
	FindOverlapping(ctx context.Context, unitTypeID string, start, end string) ([]*PriceRule, error)
	FindByUnitType(ctx context.Context, unitTypeID string) ([]*PriceRule, error)
	FindByPropertyID(ctx context.Context, propertyID string) ([]*PriceRule, error)
	Delete(ctx context.Context, id string) error
}
