package domain

import (
	"testing"
	"time"
)

func TestAvailabilityReadModel(t *testing.T) {
	date := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	model := AvailabilityReadModel{
		PropertyID: "prop-123",
		UnitID:     "unit-456",
		Date:       date,
		Count:      5,
	}
	
	if model.PropertyID != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", model.PropertyID)
	}
	if model.UnitID != "unit-456" {
		t.Errorf("Expected UnitID 'unit-456', got '%s'", model.UnitID)
	}
	if !model.Date.Equal(date) {
		t.Errorf("Expected Date %v, got %v", date, model.Date)
	}
	if model.Count != 5 {
		t.Errorf("Expected Count 5, got %d", model.Count)
	}
}

func TestSearchCriteria(t *testing.T) {
	startDate := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 6, 20, 0, 0, 0, 0, time.UTC)
	
	criteria := SearchCriteria{
		PropertyID: "prop-123",
		StartDate:  startDate,
		EndDate:    endDate,
	}
	
	if criteria.PropertyID != "prop-123" {
		t.Errorf("Expected PropertyID 'prop-123', got '%s'", criteria.PropertyID)
	}
	if !criteria.StartDate.Equal(startDate) {
		t.Errorf("Expected StartDate %v, got %v", startDate, criteria.StartDate)
	}
	if !criteria.EndDate.Equal(endDate) {
		t.Errorf("Expected EndDate %v, got %v", endDate, criteria.EndDate)
	}
}
