package domain

import (
	"github.com/google/uuid"
	"time"
)

type UnitStatus string

const (
	UnitStatusClean       UnitStatus = "clean"
	UnitStatusDirty       UnitStatus = "dirty"
	UnitStatusOccupied    UnitStatus = "occupied"
	UnitStatusMaintenance UnitStatus = "maintenance"
)

type Unit struct {
	id         string
	propertyID string
	unitTypeID string
	name       string
	status     UnitStatus
	createdAt  time.Time
}

func NewUnit(propertyID, unitTypeID, name string) *Unit {
	return &Unit{
		id:         uuid.New().String(),
		propertyID: propertyID,
		unitTypeID: unitTypeID,
		name:       name,
		status:     UnitStatusClean,
		createdAt:  time.Now(),
	}
}
func ReconstituteUnit(
	id, propertyID, unitTypeID, name, status string,
	createdAt time.Time,
) *Unit {
	return &Unit{
		id:         id,
		propertyID: propertyID,
		unitTypeID: unitTypeID,
		name:       name,
		status:     UnitStatus(status),
		createdAt:  createdAt,
	}
}
func (u *Unit) ID() string         { return u.id }
func (u *Unit) PropertyID() string { return u.propertyID }
func (u *Unit) UnitTypeID() string { return u.unitTypeID }
func (u *Unit) Name() string       { return u.name }
func (u *Unit) Status() UnitStatus { return u.status }
func (u *Unit) CheckIn()           { u.status = UnitStatusOccupied }
func (u *Unit) CheckOut()          { u.status = UnitStatusDirty }
func (u *Unit) Clean()             { u.status = UnitStatusClean }
func (u *Unit) Update(name string, status UnitStatus) {
	if name != "" {
		u.name = name
	}
	if status != "" {
		u.status = status
	}
}
