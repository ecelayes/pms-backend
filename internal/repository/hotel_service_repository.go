package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type HotelServiceRepository struct {
	db *pgxpool.Pool
}

func NewHotelServiceRepository(db *pgxpool.Pool) *HotelServiceRepository {
	return &HotelServiceRepository{db: db}
}

func (r *HotelServiceRepository) Create(ctx context.Context, s entity.HotelService) error {
	query := `
		INSERT INTO hotel_services (id, name, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, s.ID, s.Name, s.Description, s.Icon)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrConflict
		}
		return fmt.Errorf("create hotel service: %w", err)
	}
	return nil
}

func (r *HotelServiceRepository) GetAll(ctx context.Context, pagination entity.PaginationRequest) ([]entity.HotelService, int64, error) {
	countQuery := `SELECT COUNT(*) FROM hotel_services WHERE deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count hotel services: %w", err)
	}

	var query string
	var args []interface{}

	if pagination.Unlimited {
		query = `
			SELECT id, name, description, icon 
			FROM hotel_services 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
		`
	} else {
		query = `
			SELECT id, name, description, icon 
			FROM hotel_services 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
			LIMIT $1 OFFSET $2
		`
		offset := (pagination.Page - 1) * pagination.Limit
		args = append(args, pagination.Limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list hotel services: %w", err)
	}
	defer rows.Close()

	var list []entity.HotelService
	for rows.Next() {
		var s entity.HotelService
		if err := rows.Scan(&s.ID, &s.Name, &s.Description, &s.Icon); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, nil
}

func (r *HotelServiceRepository) GetByID(ctx context.Context, id string) (*entity.HotelService, error) {
	query := `SELECT id, name, description, icon FROM hotel_services WHERE id = $1 AND deleted_at IS NULL`
	var s entity.HotelService
	err := r.db.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.Description, &s.Icon)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get service by id: %w", err)
	}
	return &s, nil
}

func (r *HotelServiceRepository) Update(ctx context.Context, id string, req entity.UpdateCatalogRequest) error {
	query := `
		UPDATE hotel_services 
		SET name = $2, description = $3, icon = $4, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, id, req.Name, req.Description, req.Icon)
	if err != nil {
		return fmt.Errorf("update hotel service: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *HotelServiceRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE hotel_services SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete hotel service: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}
