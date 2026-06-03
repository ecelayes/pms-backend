package domain

import (
	"errors"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/google/uuid"
	"time"
)

var ErrInvalidUnitType = errors.New("invalid unit type")

type UnitType struct {
	id            string
	propertyID    string
	name          string
	code          string
	totalQuantity int
	basePrice     vo.Money
	maxOccupancy  int
	maxAdults     int
	maxChildren   int
	amenities     []string
	createdAt     time.Time
}

func NewUnitType(
	propertyID, name, code string,
	totalQty int,
	basePrice vo.Money,
	maxOcc, maxAdults, maxChildren int,
	amenities []string,
) (*UnitType, error) {
	if propertyID == "" || name == "" || code == "" {
		return nil, ErrInvalidUnitType
	}
	if maxOcc < 1 {
		return nil, ErrInvalidUnitType
	}
	if basePrice.Amount() < 0 {
		return nil, ErrInvalidUnitType
	}
	return &UnitType{
		id:            uuid.New().String(),
		propertyID:    propertyID,
		name:          name,
		code:          code,
		totalQuantity: totalQty,
		basePrice:     basePrice,
		maxOccupancy:  maxOcc,
		maxAdults:     maxAdults,
		maxChildren:   maxChildren,
		amenities:     amenities,
		createdAt:     time.Now(),
	}, nil
}
func ReconstituteUnitType(
	id, propertyID, name, code string,
	totalQty int,
	basePrice vo.Money,
	maxOcc, maxAdults, maxChildren int,
	amenities []string,
	createdAt time.Time,
) *UnitType {
	return &UnitType{
		id:            id,
		propertyID:    propertyID,
		name:          name,
		code:          code,
		totalQuantity: totalQty,
		basePrice:     basePrice,
		maxOccupancy:  maxOcc,
		maxAdults:     maxAdults,
		maxChildren:   maxChildren,
		amenities:     amenities,
		createdAt:     createdAt,
	}
}
func (ut *UnitType) ID() string          { return ut.id }
func (ut *UnitType) PropertyID() string  { return ut.propertyID }
func (ut *UnitType) Name() string        { return ut.name }
func (ut *UnitType) Code() string        { return ut.code }
func (ut *UnitType) TotalQuantity() int  { return ut.totalQuantity }
func (ut *UnitType) BasePrice() vo.Money { return ut.basePrice }
func (ut *UnitType) MaxOccupancy() int   { return ut.maxOccupancy }
func (ut *UnitType) MaxAdults() int      { return ut.maxAdults }
func (ut *UnitType) MaxChildren() int    { return ut.maxChildren }
func (ut *UnitType) Amenities() []string { return ut.amenities }
func (ut *UnitType) Update(name, code string, qty int, price vo.Money, maxOcc, maxAd, maxCh int, amenities []string) {
	if name != "" {
		ut.name = name
	}
	if code != "" {
		ut.code = code
	}
	if qty >= 0 {
		ut.totalQuantity = qty
	}
	ut.basePrice = price
	if maxOcc > 0 {
		ut.maxOccupancy = maxOcc
	}
	if maxAd > 0 {
		ut.maxAdults = maxAd
	}
	if maxCh >= 0 {
		ut.maxChildren = maxCh
	}
	if amenities != nil {
		ut.amenities = amenities
	}
}
