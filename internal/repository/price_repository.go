package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type PriceRepository struct {
	db *pgxpool.Pool
}

func NewPriceRepository(db *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{db: db}
}

func (r *PriceRepository) GetOverlapping(ctx context.Context, tx pgx.Tx, unitTypeID string, start, end time.Time) ([]entity.PriceRule, error) {
	query := `
		SELECT id, unit_type_id, LOWER(validity_range), UPPER(validity_range), price
		FROM price_rules
		WHERE unit_type_id = $1 
		  AND deleted_at IS NULL
		  AND validity_range && daterange($2::date, $3::date)
		FOR UPDATE
	`
	rows, err := tx.Query(ctx, query, unitTypeID, start, end)
	if err != nil {
		return nil, fmt.Errorf("find overlapping: %w", err)
	}
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		if err := rows.Scan(&pr.ID, &pr.UnitTypeID, &pr.Start, &pr.End, &pr.Price); err != nil {
			return nil, err
		}
		rules = append(rules, pr)
	}
	return rules, nil
}

func (r *PriceRepository) BatchDelete(ctx context.Context, tx pgx.Tx, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	query := `UPDATE price_rules SET deleted_at = NOW() WHERE id = ANY($1)`
	_, err := tx.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("batch delete: %w", err)
	}
	return nil
}

func (r *PriceRepository) BatchCreate(ctx context.Context, tx pgx.Tx, rules []entity.PriceRule) error {
	if len(rules) == 0 {
		return nil
	}

	var query strings.Builder
	query.WriteString("INSERT INTO price_rules (id, unit_type_id, validity_range, price, created_at, updated_at) VALUES ")
	
	values := make([]interface{}, 0, len(rules)*4)
	now := time.Now()

	for i, rule := range rules {
		if i > 0 {
			query.WriteString(",")
		}
		offset := i * 6
		query.WriteString(fmt.Sprintf("($%d, $%d, $%d::daterange, $%d, $%d, $%d)", 
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6))

		rangeStr := fmt.Sprintf("[%s, %s)", rule.Start.Format("2006-01-02"), rule.End.Format("2006-01-02"))
		values = append(values, rule.ID, rule.UnitTypeID, rangeStr, rule.Price, now, now)
	}

	_, err := tx.Exec(ctx, query.String(), values...)
	if err != nil {
		return fmt.Errorf("batch create: %w", err)
	}
	return nil
}

func (r *PriceRepository) ListByUnitType(ctx context.Context, unitTypeID string, pagination entity.PaginationRequest) ([]entity.PriceRule, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM price_rules
		WHERE unit_type_id = $1 AND deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, unitTypeID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count price rules: %w", err)
	}

	query := `
		SELECT id, unit_type_id, LOWER(validity_range), UPPER(validity_range), price, created_at, updated_at
		FROM price_rules
		WHERE unit_type_id = $1 AND deleted_at IS NULL
		ORDER BY validity_range ASC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, unitTypeID, pagination.Limit, offset)
	if err != nil { return nil, 0, fmt.Errorf("list: %w", err) }
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		if err := rows.Scan(&pr.ID, &pr.UnitTypeID, &pr.Start, &pr.End, &pr.Price, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rules = append(rules, pr)
	}
	return rules, total, nil
}

func (r *PriceRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE price_rules SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete price rule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *PriceRepository) GetByID(ctx context.Context, id string) (*entity.PriceRule, error) {
	query := `SELECT id, unit_type_id, LOWER(validity_range), UPPER(validity_range), price FROM price_rules WHERE id=$1 AND deleted_at IS NULL`
	var pr entity.PriceRule
	err := r.db.QueryRow(ctx, query, id).Scan(&pr.ID, &pr.UnitTypeID, &pr.Start, &pr.End, &pr.Price)
	if err != nil { 
		if err == pgx.ErrNoRows { return nil, entity.ErrRecordNotFound }
		return nil, err 
	}
	return &pr, nil
}

func (r *PriceRepository) ListByProperty(ctx context.Context, propertyID string, pagination entity.PaginationRequest) ([]entity.PriceRule, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM price_rules pr
		JOIN unit_types ut ON pr.unit_type_id = ut.id
		WHERE ut.property_id = $1 AND ut.deleted_at IS NULL AND pr.deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, propertyID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count price rules: %w", err)
	}

	query := `
		SELECT pr.id, pr.unit_type_id, LOWER(pr.validity_range), UPPER(pr.validity_range), pr.price, pr.created_at, pr.updated_at
		FROM price_rules pr
		JOIN unit_types ut ON pr.unit_type_id = ut.id
		WHERE ut.property_id = $1 AND ut.deleted_at IS NULL AND pr.deleted_at IS NULL
		ORDER BY pr.validity_range ASC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, propertyID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list by property: %w", err)
	}
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		if err := rows.Scan(&pr.ID, &pr.UnitTypeID, &pr.Start, &pr.End, &pr.Price, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, 0, err
		}
		rules = append(rules, pr)
	}
	return rules, total, nil
}
