package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)


type MealPlan struct {
	Type        MealType `json:"type"`
	Included    bool     `json:"included"`
	PricePerPax float64  `json:"price_per_pax"`
}

type CancellationRule struct {
	HoursBeforeCheckIn int         `json:"hours_before_check_in"`
	PenaltyType        PenaltyType `json:"penalty_type"`
	PenaltyValue       float64     `json:"penalty_value"`
}

type CancellationPolicy struct {
	IsRefundable bool               `json:"is_refundable"`
	Rules        []CancellationRule `json:"rules"`
}

type PaymentPolicy struct {
	Timing        PaymentTiming `json:"timing"`
	Method        PaymentMethod `json:"method"`
	PrepayPercent float64       `json:"prepay_percent"`
}

func (m *MealPlan) Scan(value interface{}) error {
	return jsonScan(value, m)
}
func (m MealPlan) Value() (driver.Value, error) {
	return jsonValue(m)
}

func (c *CancellationPolicy) Scan(value interface{}) error {
	return jsonScan(value, c)
}
func (c CancellationPolicy) Value() (driver.Value, error) {
	return jsonValue(c)
}

func (p *PaymentPolicy) Scan(value interface{}) error {
	return jsonScan(value, p)
}
func (p PaymentPolicy) Value() (driver.Value, error) {
	return jsonValue(p)
}

func jsonScan(value interface{}, target interface{}) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to unmarshal JSONB: expected []byte or string, got %T", value)
	}

	return json.Unmarshal(bytes, target)
}

func jsonValue(item interface{}) (driver.Value, error) {
	return json.Marshal(item)
}

type RatePlan struct {
	BaseEntity
	
	PropertyID string `json:"property_id"`
	UnitTypeID *string `json:"unit_type_id,omitempty"`

	Name        string `json:"name"`
	Description string `json:"description"`
	Active      bool   `json:"active"`

	MealPlan           MealPlan           `json:"meal_plan"`
	CancellationPolicy CancellationPolicy `json:"cancellation_policy"`
	PaymentPolicy      PaymentPolicy      `json:"payment_policy"`
}

type CreateRatePlanRequest struct {
	PropertyID         string             `json:"property_id"`
	UnitTypeID         *string            `json:"unit_type_id"`
	Name               string             `json:"name"`
	Description        string             `json:"description"`
	MealPlan           MealPlan           `json:"meal_plan"`
	CancellationPolicy CancellationPolicy `json:"cancellation_policy"`
	PaymentPolicy      PaymentPolicy      `json:"payment_policy"`
}

type UpdateRatePlanRequest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Active             *bool               `json:"active"`
	MealPlan           *MealPlan           `json:"meal_plan"`
	CancellationPolicy *CancellationPolicy `json:"cancellation_policy"`
	PaymentPolicy      *PaymentPolicy      `json:"payment_policy"`
}

func (cp *CancellationPolicy) CalculatePenaltyAmount(totalPrice float64, firstNightPrice float64, hoursUntilCheckIn float64) float64 {
	if !cp.IsRefundable {
		return totalPrice
	}

	if len(cp.Rules) == 0 {
		return 0.0
	}

	var activeRule *CancellationRule
	minHours := 1000000.0

	for _, rule := range cp.Rules {
		if hoursUntilCheckIn <= float64(rule.HoursBeforeCheckIn) {
			if float64(rule.HoursBeforeCheckIn) < minHours {
				minHours = float64(rule.HoursBeforeCheckIn)
				r := rule
				activeRule = &r
			}
		}
	}

	if activeRule == nil {
		return 0.0
	}

	var penalty float64

	switch activeRule.PenaltyType {
	case PenaltyFixedAmount:
		penalty = activeRule.PenaltyValue
	case PenaltyPercentage:
		penalty = totalPrice * (activeRule.PenaltyValue / 100.0)
	case PenaltyNights:
		penalty = firstNightPrice * activeRule.PenaltyValue
	}

	if penalty > totalPrice {
		return totalPrice
	}
	
	return penalty
}
