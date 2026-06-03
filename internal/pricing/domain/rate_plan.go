package domain

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/google/uuid"
	"time"
)

var ErrInvalidRatePlan = errors.New("invalid rate plan")

type PenaltyType string

const (
	PenaltyFixedAmount PenaltyType = "fixed"
	PenaltyPercentage  PenaltyType = "percentage"
	PenaltyNights      PenaltyType = "nights"
)

type CancellationRule struct {
	HoursBeforeCheckIn int         `json:"hours_before_check_in"`
	PenaltyType        PenaltyType `json:"penalty_type"`
	PenaltyValue       int64       `json:"penalty_value"`
}
type CancellationPolicy struct {
	IsRefundable bool               `json:"is_refundable"`
	Rules        []CancellationRule `json:"rules"`
}
type PaymentPolicy struct {
	PrepayPercent int `json:"prepay_percent"`
}
type MealPlan struct {
	Included    bool  `json:"included"`
	PricePerPax int64 `json:"price_per_pax"`
	Type        int   `json:"type"`
}
type RatePlan struct {
	id                 string
	propertyID         string
	unitTypeID         *string
	name               string
	description        string
	active             bool
	mealPlan           MealPlan
	cancellationPolicy CancellationPolicy
	paymentPolicy      PaymentPolicy
	createdAt          time.Time
}

func NewRatePlan(propertyID, name string, mealPlan MealPlan, cancel CancellationPolicy, pay PaymentPolicy) (*RatePlan, error) {
	if propertyID == "" || name == "" {
		return nil, ErrInvalidRatePlan
	}
	if !cancel.IsRefundable && len(cancel.Rules) > 0 {
		return nil, ErrInvalidRatePlan
	}
	return &RatePlan{
		id:                 uuid.New().String(),
		propertyID:         propertyID,
		name:               name,
		active:             true,
		mealPlan:           mealPlan,
		cancellationPolicy: cancel,
		paymentPolicy:      pay,
		createdAt:          time.Now(),
	}, nil
}
func ReconstituteRatePlan(
	id, propertyID string, unitTypeID *string,
	name, description string, active bool,
	mealPlan MealPlan,
	cancellationPolicy CancellationPolicy,
	paymentPolicy PaymentPolicy,
	createdAt time.Time,
) *RatePlan {
	return &RatePlan{
		id:                 id,
		propertyID:         propertyID,
		unitTypeID:         unitTypeID,
		name:               name,
		description:        description,
		active:             active,
		mealPlan:           mealPlan,
		cancellationPolicy: cancellationPolicy,
		paymentPolicy:      paymentPolicy,
		createdAt:          createdAt,
	}
}
func (m *MealPlan) Value() (driver.Value, error) {
	return json.Marshal(m)
}
func (m *MealPlan) Scan(value interface{}) error {
	if value == nil {
		*m = MealPlan{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return errors.New("type assertion failed")
	}
}
func (c *CancellationPolicy) Value() (driver.Value, error) {
	return json.Marshal(c)
}
func (c *CancellationPolicy) Scan(value interface{}) error {
	if value == nil {
		*c = CancellationPolicy{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return errors.New("type assertion failed")
	}
}
func (p *PaymentPolicy) Value() (driver.Value, error) {
	return json.Marshal(p)
}
func (p *PaymentPolicy) Scan(value interface{}) error {
	if value == nil {
		*p = PaymentPolicy{}
		return nil
	}
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, p)
	case string:
		return json.Unmarshal([]byte(v), p)
	default:
		return errors.New("type assertion failed")
	}
}
func (r *RatePlan) MealPlan() MealPlan { return r.mealPlan }
func (r *RatePlan) CalculateTotal(basePrice vo.Money, start, end time.Time, adults, children int) vo.Money {
	days := int(end.Sub(start).Hours() / 24)
	if days < 1 {
		days = 1
	}
	total := basePrice
	if r.mealPlan.Included || r.mealPlan.PricePerPax > 0 {
		pax := int64(adults + children)
		markupAmount := r.mealPlan.PricePerPax * pax * int64(days)
		markup := vo.NewMoney(markupAmount, basePrice.Currency())
		t, _ := total.Add(markup)
		total = t
	}
	return total
}
func (r *RatePlan) ID() string                             { return r.id }
func (r *RatePlan) PropertyID() string                     { return r.propertyID }
func (r *RatePlan) UnitTypeID() *string                    { return r.unitTypeID }
func (r *RatePlan) Name() string                           { return r.name }
func (r *RatePlan) Description() string                    { return r.description }
func (r *RatePlan) Active() bool                           { return r.active }
func (r *RatePlan) CancellationPolicy() CancellationPolicy { return r.cancellationPolicy }
func (r *RatePlan) PaymentPolicy() PaymentPolicy           { return r.paymentPolicy }
func (r *RatePlan) Update(name, description string, active bool) {
	if name != "" {
		r.name = name
	}
	if description != "" {
		r.description = description
	}
	r.active = active
}
