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

func (r *PriceRepository) GetOverlapping(ctx context.Context, tx pgx.Tx, roomTypeID string, start, end time.Time) ([]entity.PriceRule, error) {
	query := `
		SELECT id, room_type_id, LOWER(validity_range), UPPER(validity_range), price
		FROM price_rules
		WHERE room_type_id = $1 
		  AND deleted_at IS NULL
		  AND validity_range && daterange($2::date, $3::date)
		FOR UPDATE
	`
	rows, err := tx.Query(ctx, query, roomTypeID, start, end)
	if err != nil {
		return nil, fmt.Errorf("find overlapping: %w", err)
	}
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		if err := rows.Scan(&pr.ID, &pr.RoomTypeID, &pr.Start, &pr.End, &pr.Price); err != nil {
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

	query := "INSERT INTO price_rules (id, room_type_id, validity_range, price, created_at, updated_at) VALUES "
	values := []interface{}{}
	placeholders := []string{}
	
	for i, rule := range rules {
		offset := i * 4
		rangeStr := fmt.Sprintf("[%s, %s)", rule.Start.Format("2006-01-02"), rule.End.Format("2006-01-02"))
		
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d::daterange, $%d, NOW(), NOW())", 
			offset+1, offset+2, offset+3, offset+4))
		
		values = append(values, rule.ID, rule.RoomTypeID, rangeStr, rule.Price)
	}

	query += strings.Join(placeholders, ",")
	
	_, err := tx.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("batch create: %w", err)
	}
	return nil
}

func (r *PriceRepository) ListByRoomType(ctx context.Context, roomTypeID string) ([]entity.PriceRule, error) {
	query := `
		SELECT id, room_type_id, LOWER(validity_range), UPPER(validity_range), price, created_at, updated_at
		FROM price_rules
		WHERE room_type_id = $1 AND deleted_at IS NULL
		ORDER BY validity_range ASC
	`
	rows, err := r.db.Query(ctx, query, roomTypeID)
	if err != nil { return nil, fmt.Errorf("list: %w", err) }
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		if err := rows.Scan(&pr.ID, &pr.RoomTypeID, &pr.Start, &pr.End, &pr.Price, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, pr)
	}
	return rules, nil
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
	query := `SELECT id, room_type_id, LOWER(validity_range), UPPER(validity_range), price FROM price_rules WHERE id=$1 AND deleted_at IS NULL`
	var pr entity.PriceRule
	err := r.db.QueryRow(ctx, query, id).Scan(&pr.ID, &pr.RoomTypeID, &pr.Start, &pr.End, &pr.Price)
	if err != nil { 
		if err == pgx.ErrNoRows { return nil, entity.ErrRecordNotFound }
		return nil, err 
	}
	return &pr, nil
}
