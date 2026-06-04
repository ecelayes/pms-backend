package domain

import (
	"testing"
	"time"

	"github.com/ecelayes/pms-backend/internal/shared/vo"
)

func TestNewUnitType(t *testing.T) {
	price := vo.NewMoney(15000, "USD")
	ut, err := NewUnitType(
		"prop-123", "Deluxe Room", "DR",
		10, price, 4, 2, 2,
		[]string{"wifi", "tv"},
	)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if ut.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if ut.PropertyID() != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", ut.PropertyID())
	}
	if ut.Name() != "Deluxe Room" {
		t.Errorf("Expected Name 'Deluxe Room', got '%s'", ut.Name())
	}
	if ut.Code() != "DR" {
		t.Errorf("Expected Code 'DR', got '%s'", ut.Code())
	}
	if ut.TotalQuantity() != 10 {
		t.Errorf("Expected TotalQuantity 10, got %d", ut.TotalQuantity())
	}
	if ut.BasePrice().Amount() != 15000 {
		t.Errorf("Expected BasePrice 15000, got %d", ut.BasePrice().Amount())
	}
	if ut.MaxOccupancy() != 4 {
		t.Errorf("Expected MaxOccupancy 4, got %d", ut.MaxOccupancy())
	}
	if ut.MaxAdults() != 2 {
		t.Errorf("Expected MaxAdults 2, got %d", ut.MaxAdults())
	}
	if ut.MaxChildren() != 2 {
		t.Errorf("Expected MaxChildren 2, got %d", ut.MaxChildren())
	}
	if len(ut.Amenities()) != 2 {
		t.Errorf("Expected 2 amenities, got %d", len(ut.Amenities()))
	}
}

func TestNewUnitTypeEmptyPropertyID(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, err := NewUnitType(
		"", "Standard Room", "SR",
		5, price, 2, 2, 0,
		[]string{},
	)
	
	if err != ErrInvalidUnitType {
		t.Errorf("Expected ErrInvalidUnitType, got %v", err)
	}
	if ut != nil {
		t.Error("Expected nil UnitType on invalid propertyID")
	}
}

func TestNewUnitTypeEmptyName(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, err := NewUnitType(
		"prop-123", "", "SR",
		5, price, 2, 2, 0,
		[]string{},
	)
	
	if err != ErrInvalidUnitType {
		t.Errorf("Expected ErrInvalidUnitType, got %v", err)
	}
	if ut != nil {
		t.Error("Expected nil UnitType on invalid name")
	}
}

func TestNewUnitTypeEmptyCode(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, err := NewUnitType(
		"prop-123", "Standard Room", "",
		5, price, 2, 2, 0,
		[]string{},
	)
	
	if err != ErrInvalidUnitType {
		t.Errorf("Expected ErrInvalidUnitType, got %v", err)
	}
	if ut != nil {
		t.Error("Expected nil UnitType on invalid code")
	}
}

func TestNewUnitTypeInvalidOccupancy(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, err := NewUnitType(
		"prop-123", "Standard Room", "SR",
		5, price, 0, 2, 0, // maxOcc = 0 is invalid
		[]string{},
	)
	
	if err != ErrInvalidUnitType {
		t.Errorf("Expected ErrInvalidUnitType, got %v", err)
	}
	if ut != nil {
		t.Error("Expected nil UnitType on invalid maxOccupancy")
	}
}

func TestNewUnitTypeNegativePrice(t *testing.T) {
	price := vo.NewMoney(-100, "USD")
	ut, err := NewUnitType(
		"prop-123", "Standard Room", "SR",
		5, price, 2, 2, 0,
		[]string{},
	)
	
	if err != ErrInvalidUnitType {
		t.Errorf("Expected ErrInvalidUnitType, got %v", err)
	}
	if ut != nil {
		t.Error("Expected nil UnitType on negative price")
	}
}

func TestReconstituteUnitType(t *testing.T) {
	price := vo.NewMoney(20000, "EUR")
	createdAt := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	
	ut := ReconstituteUnitType(
		"ut-456", "prop-789", "Suite", "S",
		5, price, 6, 3, 3,
		[]string{"wifi", "mini-bar", "jacuzzi"},
		createdAt,
	)
	
	if ut.ID() != "ut-456" {
		t.Errorf("Expected ID 'ut-456', got '%s'", ut.ID())
	}
	if ut.PropertyID() != "prop-789" {
		t.Errorf("Expected PropertyID 'prop-789', got '%s'", ut.PropertyID())
	}
	if ut.Name() != "Suite" {
		t.Errorf("Expected Name 'Suite', got '%s'", ut.Name())
	}
	if ut.Code() != "S" {
		t.Errorf("Expected Code 'S', got '%s'", ut.Code())
	}
	if ut.TotalQuantity() != 5 {
		t.Errorf("Expected TotalQuantity 5, got %d", ut.TotalQuantity())
	}
	if ut.BasePrice().Currency() != "EUR" {
		t.Errorf("Expected Currency EUR, got %s", ut.BasePrice().Currency())
	}
	if ut.MaxOccupancy() != 6 {
		t.Errorf("Expected MaxOccupancy 6, got %d", ut.MaxOccupancy())
	}
	if len(ut.Amenities()) != 3 {
		t.Errorf("Expected 3 amenities, got %d", len(ut.Amenities()))
	}
}

func TestUnitTypeGetters(t *testing.T) {
	price := vo.NewMoney(12000, "GBP")
	ut, _ := NewUnitType(
		"prop-123", "Standard", "STD",
		20, price, 2, 2, 0,
		[]string{"wifi"},
	)
	
	if ut.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if ut.PropertyID() != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", ut.PropertyID())
	}
	if ut.Name() != "Standard" {
		t.Errorf("Expected Name 'Standard', got '%s'", ut.Name())
	}
	if ut.Code() != "STD" {
		t.Errorf("Expected Code 'STD', got '%s'", ut.Code())
	}
	if ut.TotalQuantity() != 20 {
		t.Errorf("Expected TotalQuantity 20, got %d", ut.TotalQuantity())
	}
	if ut.BasePrice().Amount() != 12000 {
		t.Errorf("Expected BasePrice 12000, got %d", ut.BasePrice().Amount())
	}
	if ut.MaxOccupancy() != 2 {
		t.Errorf("Expected MaxOccupancy 2, got %d", ut.MaxOccupancy())
	}
	if ut.MaxAdults() != 2 {
		t.Errorf("Expected MaxAdults 2, got %d", ut.MaxAdults())
	}
	if ut.MaxChildren() != 0 {
		t.Errorf("Expected MaxChildren 0, got %d", ut.MaxChildren())
	}
}

func TestUnitTypeUpdate(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, _ := NewUnitType(
		"prop-123", "Old Name", "OLD",
		5, price, 2, 2, 0,
		[]string{"wifi"},
	)
	newPrice := vo.NewMoney(15000, "USD")
	
	ut.Update("New Name", "NEW", 10, newPrice, 4, 3, 2, []string{"wifi", "tv", "ac"})
	
	if ut.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", ut.Name())
	}
	if ut.Code() != "NEW" {
		t.Errorf("Expected Code 'NEW', got '%s'", ut.Code())
	}
	if ut.TotalQuantity() != 10 {
		t.Errorf("Expected TotalQuantity 10, got %d", ut.TotalQuantity())
	}
	if ut.BasePrice().Amount() != 15000 {
		t.Errorf("Expected BasePrice 15000, got %d", ut.BasePrice().Amount())
	}
	if ut.MaxOccupancy() != 4 {
		t.Errorf("Expected MaxOccupancy 4, got %d", ut.MaxOccupancy())
	}
	if ut.MaxAdults() != 3 {
		t.Errorf("Expected MaxAdults 3, got %d", ut.MaxAdults())
	}
	if ut.MaxChildren() != 2 {
		t.Errorf("Expected MaxChildren 2, got %d", ut.MaxChildren())
	}
	if len(ut.Amenities()) != 3 {
		t.Errorf("Expected 3 amenities, got %d", len(ut.Amenities()))
	}
}

func TestUnitTypeUpdatePartial(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, _ := NewUnitType(
		"prop-123", "Original", "ORIG",
		5, price, 2, 2, 0,
		[]string{"wifi"},
	)
	originalName := ut.Name()
	originalCode := ut.Code()
	originalQty := ut.TotalQuantity()
	
	ut.Update("", "", -1, vo.Money{}, 0, 0, 0, nil)
	
	if ut.Name() != originalName {
		t.Error("Name should not change when empty string passed")
	}
	if ut.Code() != originalCode {
		t.Error("Code should not change when empty string passed")
	}
	if ut.TotalQuantity() != originalQty {
		t.Error("TotalQuantity should not change when -1 passed")
	}
}

func TestUnitTypeAmenities(t *testing.T) {
	price := vo.NewMoney(10000, "USD")
	ut, _ := NewUnitType(
		"prop-123", "Room", "R",
		5, price, 2, 2, 0,
		[]string{"wifi", "tv", "safe", "minibar"},
	)
	
	amenities := ut.Amenities()
	if len(amenities) != 4 {
		t.Errorf("Expected 4 amenities, got %d", len(amenities))
	}
	
	if amenities[0] != "wifi" {
		t.Errorf("Expected first amenity 'wifi', got '%s'", amenities[0])
	}
	if amenities[3] != "minibar" {
		t.Errorf("Expected fourth amenity 'minibar', got '%s'", amenities[3])
	}
}
