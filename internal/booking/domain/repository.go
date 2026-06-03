package domain

import (
	"context"
	"time"
)

type ReservationRepository interface {
	Save(ctx context.Context, reservation *Reservation) error
	FindByID(ctx context.Context, id string) (*Reservation, error)
	Update(ctx context.Context, reservation *Reservation) error
	CountOverlapping(ctx context.Context, unitTypeID string, start, end time.Time) (int, error)
	GetByCode(ctx context.Context, code string) (*Reservation, error)
	LockUnitType(ctx context.Context, unitTypeID string) error
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
