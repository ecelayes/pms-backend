package domain

import (
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/google/uuid"
	"time"
)

type PriceRule struct {
	id         string
	unitTypeID string
	dateRange  vo.DateRange
	price      vo.Money
	createdAt  time.Time
}

func NewPriceRule(unitTypeID string, dateRange vo.DateRange, price vo.Money) *PriceRule {
	return &PriceRule{
		id:         uuid.New().String(),
		unitTypeID: unitTypeID,
		dateRange:  dateRange,
		price:      price,
		createdAt:  time.Now(),
	}
}
func ReconstitutePriceRule(
	id, unitTypeID string,
	dateRange vo.DateRange,
	price vo.Money,
	createdAt time.Time,
) *PriceRule {
	return &PriceRule{
		id:         id,
		unitTypeID: unitTypeID,
		dateRange:  dateRange,
		price:      price,
		createdAt:  createdAt,
	}
}
func (p *PriceRule) ID() string              { return p.id }
func (p *PriceRule) UnitTypeID() string      { return p.unitTypeID }
func (p *PriceRule) DateRange() vo.DateRange { return p.dateRange }
func (p *PriceRule) Price() vo.Money         { return p.price }
