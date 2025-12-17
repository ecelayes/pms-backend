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

type AmenityRepository struct {
	db *pgxpool.Pool
}

func NewAmenityRepository(db *pgxpool.Pool) *AmenityRepository {
	return &AmenityRepository{db: db}
}

func (r *AmenityRepository) Create(ctx context.Context, a entity.Amenity) error {
	query := `INSERT INTO amenities (id, name, description, icon, created_at, updated_at) VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err := r.db.Exec(ctx, query, a.ID, a.Name, a.Description, a.Icon)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.ErrConflict
		}
		return err
	}
	return nil
}

func (r *AmenityRepository) GetAll(ctx context.Context, pagination entity.PaginationRequest) ([]entity.Amenity, int64, error) {
	countQuery := `SELECT COUNT(*) FROM amenities WHERE deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count amenities: %w", err)
	}

	var query string
	var args []interface{}

	if pagination.Unlimited {
		query = `
			SELECT id, name, description, icon 
			FROM amenities 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
		`
	} else {
		query = `
			SELECT id, name, description, icon 
			FROM amenities 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
			LIMIT $1 OFFSET $2
		`
		offset := (pagination.Page - 1) * pagination.Limit
		args = append(args, pagination.Limit, offset)
	}
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil { return nil, 0, err }
	defer rows.Close()

	var list []entity.Amenity
	for rows.Next() {
		var a entity.Amenity
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon); err != nil { return nil, 0, err }
		list = append(list, a)
	}
	return list, total, nil
}

func (r *AmenityRepository) GetByID(ctx context.Context, id string) (*entity.Amenity, error) {
	query := `SELECT id, name, description, icon FROM amenities WHERE id = $1 AND deleted_at IS NULL`
	var a entity.Amenity
	err := r.db.QueryRow(ctx, query, id).Scan(&a.ID, &a.Name, &a.Description, &a.Icon)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get amenity by id: %w", err)
	}
	return &a, nil
}

func (r *AmenityRepository) Update(ctx context.Context, id string, req entity.UpdateCatalogRequest) error {
	query := `UPDATE amenities SET name=$2, description=$3, icon=$4, updated_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id, req.Name, req.Description, req.Icon)
	if err != nil {
		return fmt.Errorf("update amenity: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *AmenityRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE amenities SET deleted_at=NOW() WHERE id=$1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete amenity: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}
