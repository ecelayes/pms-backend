package domain

import (
	"time"
	"testing"
)

func TestNewOrganization(t *testing.T) {
	org, err := NewOrganization("Acme Corp", "ACME")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if org.ID() == "" {
		t.Error("Expected non-empty ID")
	}
	if org.Name() != "Acme Corp" {
		t.Errorf("Expected Name 'Acme Corp', got '%s'", org.Name())
	}
	if org.Code() != "ACME" {
		t.Errorf("Expected Code 'ACME', got '%s'", org.Code())
	}
}

func TestNewOrganizationEmptyName(t *testing.T) {
	org, err := NewOrganization("", "ACME")
	
	if err != ErrInvalidOrganization {
		t.Errorf("Expected ErrInvalidOrganization, got %v", err)
	}
	if org != nil {
		t.Error("Expected nil organization on invalid name")
	}
}

func TestNewOrganizationEmptyCode(t *testing.T) {
	org, err := NewOrganization("Acme Corp", "")
	
	if err != ErrInvalidOrganization {
		t.Errorf("Expected ErrInvalidOrganization, got %v", err)
	}
	if org != nil {
		t.Error("Expected nil organization on invalid code")
	}
}

func TestOrganizationUpdate(t *testing.T) {
	org, _ := NewOrganization("Old Name", "OLD")
	
	org.Update("New Name", "NEW")
	
	if org.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", org.Name())
	}
	if org.Code() != "NEW" {
		t.Errorf("Expected Code 'NEW', got '%s'", org.Code())
	}
}

func TestOrganizationUpdatePartialName(t *testing.T) {
	org, _ := NewOrganization("Original Name", "ORIG")
	originalCode := org.Code()
	
	org.Update("", "NEW")
	
	if org.Name() != "Original Name" {
		t.Errorf("Name should not change when empty string passed")
	}
	if org.Code() != "NEW" {
		t.Errorf("Expected Code 'NEW', got '%s'", org.Code())
	}
	_ = originalCode // silence unused warning
}

func TestOrganizationUpdatePartialCode(t *testing.T) {
	org, _ := NewOrganization("Original Name", "ORIG")
	originalName := org.Name()
	
	org.Update("New Name", "")
	
	if org.Name() != "New Name" {
		t.Errorf("Expected Name 'New Name', got '%s'", org.Name())
	}
	if org.Code() != "ORIG" {
		t.Errorf("Code should not change when empty string passed")
	}
	_ = originalName // silence unused warning
}

func TestReconstituteOrganization(t *testing.T) {
	org := ReconstituteOrganization("org-123", "Test Corp", "TEST", time.Now())
	
	if org.ID() != "org-123" {
		t.Errorf("Expected ID 'org-123', got '%s'", org.ID())
	}
	if org.Name() != "Test Corp" {
		t.Errorf("Expected Name 'Test Corp', got '%s'", org.Name())
	}
	if org.Code() != "TEST" {
		t.Errorf("Expected Code 'TEST', got '%s'", org.Code())
	}
}
