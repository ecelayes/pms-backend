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
	query := `
		INSERT INTO hotels (owner_id, name, code)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, ownerID, req.Name, req.Code).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create hotel: %w", err)
	}
	return id, nil
}
