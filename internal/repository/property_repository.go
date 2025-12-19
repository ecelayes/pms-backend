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

type PropertyRepository struct {
	db *pgxpool.Pool
}

func NewPropertyRepository(db *pgxpool.Pool) *PropertyRepository {
	return &PropertyRepository{db: db}
}

func (r *PropertyRepository) Create(ctx context.Context, p entity.Property) (string, error) {
	query := `
		INSERT INTO properties (id, organization_id, name, code, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, p.ID, p.OrganizationID, p.Name, p.Code, p.Type).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", entity.ErrConflict
		}
		return "", fmt.Errorf("create property: %w", err)
	}
	return id, nil
}

func (r *PropertyRepository) ListByOrganization(ctx context.Context, orgID string, pagination entity.PaginationRequest) ([]entity.Property, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM properties
		WHERE organization_id = $1 AND deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count properties: %w", err)
	}

	query := `
		SELECT id, organization_id, name, code, type, created_at, updated_at 
		FROM properties 
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, orgID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	var properties []entity.Property
	for rows.Next() {
		var p entity.Property
		if err := rows.Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Code, &p.Type, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		properties = append(properties, p)
	}
	return properties, total, nil
}

func (r *PropertyRepository) GetByID(ctx context.Context, id string) (*entity.Property, error) {
	query := `
		SELECT id, organization_id, name, code, type, created_at, updated_at 
		FROM properties 
		WHERE id = $1 AND deleted_at IS NULL
	`
	var p entity.Property
	err := r.db.QueryRow(ctx, query, id).Scan(&p.ID, &p.OrganizationID, &p.Name, &p.Code, &p.Type, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("property not found: %w", err)
	}
	return &p, nil
}

func (r *PropertyRepository) Update(ctx context.Context, id string, req entity.UpdatePropertyRequest) error {
	query := `UPDATE properties SET updated_at = NOW()`
	args := []interface{}{}
	argID := 1

	if req.Name != "" {
		query += fmt.Sprintf(", name = $%d", argID)
		args = append(args, req.Name)
		argID++
	}
	if req.Code != "" {
		query += fmt.Sprintf(", code = $%d", argID)
		args = append(args, req.Code)
		argID++
	}
	if req.Type != "" {
		query += fmt.Sprintf(", type = $%d", argID)
		args = append(args, req.Type)
		argID++
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argID)
	args = append(args, id)

	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update property: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *PropertyRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE properties SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete property: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}
