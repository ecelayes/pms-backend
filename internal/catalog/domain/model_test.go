package domain

import (
	"time"
	"testing"
)

func TestNewUnit(t *testing.T) {
	unit := NewUnit("prop-123", "unit-type-456", "Room 101")
	
	if unit.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if unit.PropertyID() != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", unit.PropertyID())
	}
	if unit.UnitTypeID() != "unit-type-456" {
		t.Errorf("Expected UnitTypeID 'unit-type-456', got '%s'", unit.UnitTypeID())
	}
	if unit.Name() != "Room 101" {
		t.Errorf("Expected Name 'Room 101', got '%s'", unit.Name())
	}
	if unit.Status() != UnitStatusClean {
		t.Errorf("Expected Status UnitStatusClean, got '%s'", unit.Status())
	}
}

func TestUnitStatusConstants(t *testing.T) {
	if UnitStatusClean != "clean" {
		t.Errorf("Expected UnitStatusClean to be 'clean', got '%s'", UnitStatusClean)
	}
	if UnitStatusDirty != "dirty" {
		t.Errorf("Expected UnitStatusDirty to be 'dirty', got '%s'", UnitStatusDirty)
	}
	if UnitStatusOccupied != "occupied" {
		t.Errorf("Expected UnitStatusOccupied to be 'occupied', got '%s'", UnitStatusOccupied)
	}
	if UnitStatusMaintenance != "maintenance" {
		t.Errorf("Expected UnitStatusMaintenance to be 'maintenance', got '%s'", UnitStatusMaintenance)
	}
}

func TestUnitCheckIn(t *testing.T) {
	unit := NewUnit("prop-123", "ut-456", "Room 101")
	unit.CheckIn()
	
	if unit.Status() != UnitStatusOccupied {
		t.Errorf("Expected Status UnitStatusOccupied after CheckIn, got '%s'", unit.Status())
	}
}

func TestUnitCheckOut(t *testing.T) {
	unit := NewUnit("prop-123", "ut-456", "Room 101")
	unit.CheckOut()
	
	if unit.Status() != UnitStatusDirty {
		t.Errorf("Expected Status UnitStatusDirty after CheckOut, got '%s'", unit.Status())
	}
}

func TestUnitClean(t *testing.T) {
	unit := NewUnit("prop-123", "ut-456", "Room 101")
	unit.CheckIn()
	unit.Clean()
	
	if unit.Status() != UnitStatusClean {
		t.Errorf("Expected Status UnitStatusClean after Clean, got '%s'", unit.Status())
	}
}

func TestUnitUpdate(t *testing.T) {
	unit := NewUnit("prop-123", "ut-456", "Room 101")
	
	unit.Update("Room 202", UnitStatusMaintenance)
	
	if unit.Name() != "Room 202" {
		t.Errorf("Expected Name 'Room 202', got '%s'", unit.Name())
	}
	if unit.Status() != UnitStatusMaintenance {
		t.Errorf("Expected Status UnitStatusMaintenance, got '%s'", unit.Status())
	}
}

func TestUnitUpdatePartial(t *testing.T) {
	unit := NewUnit("prop-123", "ut-456", "Room 101")
	originalName := unit.Name()
	
	unit.Update("", UnitStatusMaintenance)
	
	if unit.Name() != originalName {
		t.Errorf("Name should not change when empty string passed")
	}
	if unit.Status() != UnitStatusMaintenance {
		t.Errorf("Expected Status UnitStatusMaintenance, got '%s'", unit.Status())
	}
}

func TestReconstituteUnit(t *testing.T) {
	unit := ReconstituteUnit("unit-123", "prop-456", "ut-789", "Suite 1", "dirty", time.Now())
	
	if unit.ID() != "unit-123" {
		t.Errorf("Expected ID 'unit-123', got '%s'", unit.ID())
	}
	if unit.PropertyID() != "prop-456" {
		t.Errorf("Expected PropertyID 'prop-456', got '%s'", unit.PropertyID())
	}
	if unit.UnitTypeID() != "ut-789" {
		t.Errorf("Expected UnitTypeID 'ut-789', got '%s'", unit.UnitTypeID())
	}
	if unit.Name() != "Suite 1" {
		t.Errorf("Expected Name 'Suite 1', got '%s'", unit.Name())
	}
	if unit.Status() != UnitStatusDirty {
		t.Errorf("Expected Status UnitStatusDirty, got '%s'", unit.Status())
	}
}

func TestNewAmenity(t *testing.T) {
	amenity, err := NewAmenity("WiFi", "High-speed wireless", "wifi-icon")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if amenity.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if amenity.Name() != "WiFi" {
		t.Errorf("Expected Name 'WiFi', got '%s'", amenity.Name())
	}
	if amenity.Description() != "High-speed wireless" {
		t.Errorf("Expected Description 'High-speed wireless', got '%s'", amenity.Description())
	}
	if amenity.Icon() != "wifi-icon" {
		t.Errorf("Expected Icon 'wifi-icon', got '%s'", amenity.Icon())
	}
}

func TestNewAmenityEmptyName(t *testing.T) {
	amenity, err := NewAmenity("", "Description", "icon")
	
	if err != ErrInvalidName {
		t.Errorf("Expected ErrInvalidName, got %v", err)
	}
	if amenity != nil {
		t.Error("Expected nil amenity on invalid name")
	}
}

func TestAmenityUpdate(t *testing.T) {
	amenity, _ := NewAmenity("Pool", "Outdoor pool", "pool-icon")
	
	amenity.Update("Indoor Pool", "Covered swimming area", "indoor-pool")
	
	if amenity.Name() != "Indoor Pool" {
		t.Errorf("Expected Name 'Indoor Pool', got '%s'", amenity.Name())
	}
	if amenity.Description() != "Covered swimming area" {
		t.Errorf("Expected Description 'Covered swimming area', got '%s'", amenity.Description())
	}
	if amenity.Icon() != "indoor-pool" {
		t.Errorf("Expected Icon 'indoor-pool', got '%s'", amenity.Icon())
	}
}

func TestAmenityUpdatePartial(t *testing.T) {
	amenity, _ := NewAmenity("Gym", "Fitness center", "gym-icon")
	originalDesc := amenity.Description()
	
	amenity.Update("Gym", "", "")
	
	if amenity.Name() != "Gym" {
		t.Errorf("Name should not change")
	}
	if amenity.Description() != originalDesc {
		t.Errorf("Description should not change when empty")
	}
}

func TestReconstituteAmenity(t *testing.T) {
	amenity := ReconstituteAmenity("amenity-123", "Spa", "Full-service spa", "spa-icon")
	
	if amenity.ID() != "amenity-123" {
		t.Errorf("Expected ID 'amenity-123', got '%s'", amenity.ID())
	}
	if amenity.Name() != "Spa" {
		t.Errorf("Expected Name 'Spa', got '%s'", amenity.Name())
	}
}

func TestNewGuestService(t *testing.T) {
	service, err := NewGuestService("Airport Transfer", "Shuttle to airport", "car-icon")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if service.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if service.Name() != "Airport Transfer" {
		t.Errorf("Expected Name 'Airport Transfer', got '%s'", service.Name())
	}
	if service.Description() != "Shuttle to airport" {
		t.Errorf("Expected Description 'Shuttle to airport', got '%s'", service.Description())
	}
	if service.Icon() != "car-icon" {
		t.Errorf("Expected Icon 'car-icon', got '%s'", service.Icon())
	}
}

func TestNewGuestServiceEmptyName(t *testing.T) {
	service, err := NewGuestService("", "Description", "icon")
	
	if err != ErrInvalidName {
		t.Errorf("Expected ErrInvalidName, got %v", err)
	}
	if service != nil {
		t.Error("Expected nil service on invalid name")
	}
}

func TestGuestServiceUpdate(t *testing.T) {
	service, _ := NewGuestService("Breakfast", "Morning buffet", "breakfast-icon")
	
	service.Update("Brunch", "Late morning meal", "brunch-icon")
	
	if service.Name() != "Brunch" {
		t.Errorf("Expected Name 'Brunch', got '%s'", service.Name())
	}
	if service.Description() != "Late morning meal" {
		t.Errorf("Expected Description 'Late morning meal', got '%s'", service.Description())
	}
	if service.Icon() != "brunch-icon" {
		t.Errorf("Expected Icon 'brunch-icon', got '%s'", service.Icon())
	}
}

func TestGuestServiceUpdatePartial(t *testing.T) {
	service, _ := NewGuestService("WiFi", "Internet", "wifi-icon")
	originalName := service.Name()
	
	service.Update("", "Updated description", "")
	
	if service.Name() != originalName {
		t.Errorf("Name should not change when empty string passed")
	}
	if service.Description() != "Updated description" {
		t.Errorf("Expected Description 'Updated description', got '%s'", service.Description())
	}
}

func TestReconstituteGuestService(t *testing.T) {
	service := ReconstituteGuestService("service-123", "Concierge", "24/7 assistance", "concierge-icon")
	
	if service.ID() != "service-123" {
		t.Errorf("Expected ID 'service-123', got '%s'", service.ID())
	}
	if service.Name() != "Concierge" {
		t.Errorf("Expected Name 'Concierge', got '%s'", service.Name())
	}
	if service.Description() != "24/7 assistance" {
		t.Errorf("Expected Description '24/7 assistance', got '%s'", service.Description())
	}
}
