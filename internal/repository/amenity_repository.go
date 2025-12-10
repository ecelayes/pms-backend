package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	return err
}

func (r *AmenityRepository) GetAll(ctx context.Context) ([]entity.Amenity, error) {
	query := `SELECT id, name, description, icon FROM amenities WHERE deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()

	var list []entity.Amenity
	for rows.Next() {
		var a entity.Amenity
		if err := rows.Scan(&a.ID, &a.Name, &a.Description, &a.Icon); err != nil { return nil, err }
		list = append(list, a)
	}
	return list, nil
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
