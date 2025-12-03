package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type HotelRepository struct {
	db *pgxpool.Pool
}

func NewHotelRepository(db *pgxpool.Pool) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) Create(ctx context.Context, ownerID string, req entity.CreateHotelRequest) (string, error) {
	query := `INSERT INTO hotels (owner_id, name, code) VALUES ($1, $2, $3) RETURNING id`
	var id string
	err := r.db.QueryRow(ctx, query, ownerID, req.Name, req.Code).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create hotel: %w", err)
	}
	return id, nil
}

func (r *HotelRepository) GetByID(ctx context.Context, id string) (*entity.Hotel, error) {
	query := `SELECT id, owner_id, name, code, created_at, updated_at FROM hotels WHERE id = $1 AND deleted_at IS NULL`
	var h entity.Hotel
	err := r.db.QueryRow(ctx, query, id).Scan(&h.ID, &h.OwnerID, &h.Name, &h.Code, &h.CreatedAt, &h.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("hotel not found: %w", err)
	}
	return &h, nil
}

func (r *HotelRepository) Update(ctx context.Context, id string, req entity.UpdateHotelRequest) error {
	query := `UPDATE hotels SET name = $2, code = $3 WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id, req.Name, req.Code)
	if err != nil {
		return fmt.Errorf("update hotel: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("hotel not found or deleted")
	}
	return nil
}

func (r *HotelRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE hotels SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete hotel: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("hotel not found or already deleted")
	}
	return nil
}

func (r *HotelRepository) ListByOwner(ctx context.Context, ownerID string) ([]entity.Hotel, error) {
	query := `SELECT id, owner_id, name, code, created_at, updated_at FROM hotels WHERE owner_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hotels []entity.Hotel
	for rows.Next() {
		var h entity.Hotel
		if err := rows.Scan(&h.ID, &h.OwnerID, &h.Name, &h.Code, &h.CreatedAt, &h.UpdatedAt); err != nil {
			return nil, err
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}
