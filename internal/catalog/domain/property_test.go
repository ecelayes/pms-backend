package domain

import (
	"testing"
	"time"
)

func TestNewProperty(t *testing.T) {
	prop, err := NewProperty("org-123", "Beach Resort", "BR", PropertyTypeHotel)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if prop.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if prop.OrganizationID() != "org-123" {
		t.Errorf("Expected OrganizationID 'org-123', got '%s'", prop.OrganizationID())
	}
	if prop.Name() != "Beach Resort" {
		t.Errorf("Expected Name 'Beach Resort', got '%s'", prop.Name())
	}
	if prop.Code() != "BR" {
		t.Errorf("Expected Code 'BR', got '%s'", prop.Code())
	}
	if prop.Type() != PropertyTypeHotel {
		t.Errorf("Expected Type PropertyTypeHotel, got '%s'", prop.Type())
	}
}

func TestNewPropertyEmptyName(t *testing.T) {
	prop, err := NewProperty("org-123", "", "BR", PropertyTypeHotel)
	
	if err != ErrInvalidProperty {
		t.Errorf("Expected ErrInvalidProperty, got %v", err)
	}
	if prop != nil {
		t.Error("Expected nil Property on invalid name")
	}
}

func TestNewPropertyEmptyCode(t *testing.T) {
	prop, err := NewProperty("org-123", "Beach Resort", "", PropertyTypeHotel)
	
	if err != ErrInvalidProperty {
		t.Errorf("Expected ErrInvalidProperty, got %v", err)
	}
	if prop != nil {
		t.Error("Expected nil Property on invalid code")
	}
}

func TestPropertyTypeConstants(t *testing.T) {
	if PropertyTypeHotel != "hotel" {
		t.Errorf("Expected PropertyTypeHotel 'hotel', got '%s'", PropertyTypeHotel)
	}
	if PropertyTypeApartment != "apartment" {
		t.Errorf("Expected PropertyTypeApartment 'apartment', got '%s'", PropertyTypeApartment)
	}
	if PropertyTypeCabin != "cabin" {
		t.Errorf("Expected PropertyTypeCabin 'cabin', got '%s'", PropertyTypeCabin)
	}
}

func TestReconstituteProperty(t *testing.T) {
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	prop := ReconstituteProperty("prop-456", "org-789", "Mountain Lodge", "ML", "cabin", createdAt)
	
	if prop.ID() != "prop-456" {
		t.Errorf("Expected ID 'prop-456', got '%s'", prop.ID())
	}
	if prop.OrganizationID() != "org-789" {
		t.Errorf("Expected OrganizationID 'org-789', got '%s'", prop.OrganizationID())
	}
	if prop.Name() != "Mountain Lodge" {
		t.Errorf("Expected Name 'Mountain Lodge', got '%s'", prop.Name())
	}
	if prop.Code() != "ML" {
		t.Errorf("Expected Code 'ML', got '%s'", prop.Code())
	}
	if prop.Type() != PropertyTypeCabin {
		t.Errorf("Expected Type PropertyTypeCabin, got '%s'", prop.Type())
	}
}

func TestPropertyUpdate(t *testing.T) {
	prop, _ := NewProperty("org-123", "Old Name", "OLD", PropertyTypeHotel)
	
	prop.Update("New Name", "NEW", PropertyTypeApartment)
	
	if prop.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", prop.Name())
	}
	if prop.Code() != "NEW" {
		t.Errorf("Expected Code 'NEW', got '%s'", prop.Code())
	}
	if prop.Type() != PropertyTypeApartment {
		t.Errorf("Expected Type PropertyTypeApartment, got '%s'", prop.Type())
	}
}

func TestPropertyUpdatePartialName(t *testing.T) {
	prop, _ := NewProperty("org-123", "Original Name", "ORIG", PropertyTypeHotel)
	originalCode := prop.Code()
	
	prop.Update("", "NEW", PropertyTypeCabin)
	
	if prop.Name() != "Original Name" {
		t.Error("Name should not change when empty string passed")
	}
	if prop.Code() != "NEW" {
		t.Errorf("Expected Code 'NEW', got '%s'", prop.Code())
	}
	if prop.Type() != PropertyTypeCabin {
		t.Errorf("Expected Type PropertyTypeCabin, got '%s'", prop.Type())
	}
	_ = originalCode // silence
}

func TestPropertyUpdatePartialCode(t *testing.T) {
	prop, _ := NewProperty("org-123", "Original Name", "ORIG", PropertyTypeHotel)
	originalName := prop.Name()
	
	prop.Update("New Name", "", PropertyTypeCabin)
	
	if prop.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", prop.Name())
	}
	if prop.Code() != "ORIG" {
		t.Error("Code should not change when empty string passed")
	}
	if prop.Type() != PropertyTypeCabin {
		t.Errorf("Expected Type PropertyTypeCabin, got '%s'", prop.Type())
	}
	_ = originalName // silence
}

func TestPropertyUpdatePartialType(t *testing.T) {
	prop, _ := NewProperty("org-123", "Name", "CODE", PropertyTypeHotel)
	originalName := prop.Name()
	originalCode := prop.Code()
	
	prop.Update("", "", PropertyTypeCabin)
	
	if prop.Name() != originalName {
		t.Error("Name should not change")
	}
	if prop.Code() != originalCode {
		t.Error("Code should not change")
	}
	if prop.Type() != PropertyTypeCabin {
		t.Errorf("Expected Type PropertyTypeCabin, got '%s'", prop.Type())
	}
}

func TestPropertyAllPropertyTypes(t *testing.T) {
	types := []PropertyType{PropertyTypeHotel, PropertyTypeApartment, PropertyTypeCabin}
	
	for _, pt := range types {
		prop, err := NewProperty("org-123", "Test Property", "TEST", pt)
		if err != nil {
			t.Errorf("Expected no error for type %s, got %v", pt, err)
		}
		if prop.Type() != pt {
			t.Errorf("Expected Type %s, got %s", pt, prop.Type())
		}
	}
}
