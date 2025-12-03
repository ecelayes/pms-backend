package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type ReservationRepository struct {
	db *pgxpool.Pool
}

func NewReservationRepository(db *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{db: db}
}

func (r *ReservationRepository) Create(ctx context.Context, tx pgx.Tx, res entity.Reservation) error {
	query := `
		INSERT INTO reservations (id, room_type_id, reservation_code, stay_range, guest_email, total_price, status)
		VALUES ($1, $2, $3, daterange($4::date, $5::date), $6, $7, 'confirmed')
	`
	_, err := tx.Exec(ctx, query, res.ID, res.RoomTypeID, res.ReservationCode, res.Start, res.End, res.GuestEmail, res.TotalPrice)
	if err != nil {
		return fmt.Errorf("insert reservation: %w", err)
	}
	return nil
}

func (r *ReservationRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	query := `UPDATE reservations SET status = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return entity.ErrReservationNotFound
	}
	return nil
}

func (r *ReservationRepository) GetByCode(ctx context.Context, code string) (*entity.Reservation, error) {
	query := `
		SELECT id, reservation_code, room_type_id, guest_email, lower(stay_range), upper(stay_range), total_price, status, created_at, updated_at
		FROM reservations
		WHERE reservation_code = $1 AND deleted_at IS NULL
	`
	var res entity.Reservation
	err := r.db.QueryRow(ctx, query, code).Scan(
		&res.ID, &res.ReservationCode, &res.RoomTypeID, &res.GuestEmail, 
		&res.Start, &res.End, &res.TotalPrice, &res.Status, 
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("reservation not found: %w", err)
	}
	return &res, nil
}

func (r *ReservationRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE reservations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete reservation: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrReservationNotFound
	}
	return nil
}
