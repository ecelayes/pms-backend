package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UnitRepository struct {
	db *pgxpool.Pool
}

func NewUnitRepository(db *pgxpool.Pool) *UnitRepository {
	return &UnitRepository{db: db}
}

func (r *UnitRepository) Create(ctx context.Context, u entity.Unit) error {
	query := `
		INSERT INTO units (
			id, property_id, unit_type_id, name, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query,
		u.ID, u.PropertyID, u.UnitTypeID, u.Name, u.Status,
	)
	if err != nil {
		return fmt.Errorf("create unit: %w", err)
	}
	return nil
}

func (r *UnitRepository) Update(ctx context.Context, id string, req entity.UpdateUnitRequest) error {
	query := `UPDATE units SET updated_at = NOW()`
	var args []interface{}
	argID := 1
	addSet := func(column string, value interface{}) {
		query += fmt.Sprintf(", %s = $%d", column, argID)
		args = append(args, value)
		argID++
	}

	if req.Name != "" {
		addSet("name", req.Name)
	}
	if req.Status != "" {
		addSet("status", req.Status)
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argID)
	args = append(args, id)

	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update unit: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *UnitRepository) ListByProperty(ctx context.Context, propertyID string) ([]entity.Unit, error) {
	query := `
		SELECT id, property_id, unit_type_id, name, status, created_at, updated_at
		FROM units
		WHERE property_id = $1 AND deleted_at IS NULL
		ORDER BY name ASC
	`
	rows, err := r.db.Query(ctx, query, propertyID)
	if err != nil {
		return nil, fmt.Errorf("list units: %w", err)
	}
	defer rows.Close()

	var units []entity.Unit
	for rows.Next() {
		var u entity.Unit
		if err := rows.Scan(
			&u.ID, &u.PropertyID, &u.UnitTypeID, &u.Name, &u.Status, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

func (r *UnitRepository) GetByID(ctx context.Context, id string) (*entity.Unit, error) {
	query := `
		SELECT id, property_id, unit_type_id, name, status, created_at, updated_at
		FROM units
		WHERE id = $1 AND deleted_at IS NULL
	`
	var u entity.Unit
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.PropertyID, &u.UnitTypeID, &u.Name, &u.Status, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get unit: %w", err)
	}
	return &u, nil
}

func (r *UnitRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE units SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete unit: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}
