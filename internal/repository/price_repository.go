package repository

import (
	"context"
	"fmt"

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

func (r *PriceRepository) Create(ctx context.Context, rule entity.PriceRule) error {
	rangeStr := fmt.Sprintf("[%s, %s)", rule.Start.Format("2006-01-02"), rule.End.Format("2006-01-02"))
	
	query := `
		INSERT INTO price_rules (id, room_type_id, validity_range, price, priority, created_at, updated_at)
		VALUES ($1, $2, $3::daterange, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, rule.ID, rule.RoomTypeID, rangeStr, rule.Price, rule.Priority)
	if err != nil {
		return fmt.Errorf("insert price rule: %w", err)
	}
	return nil
}

func (r *PriceRepository) ListByRoomType(ctx context.Context, roomTypeID string) ([]entity.PriceRule, error) {
	query := `
		SELECT 
			id, 
			room_type_id, 
			LOWER(validity_range) as start_date, 
			UPPER(validity_range) as end_date, 
			price, 
			priority, 
			created_at, 
			updated_at
		FROM price_rules
		WHERE room_type_id = $1 AND deleted_at IS NULL
		ORDER BY priority DESC, validity_range ASC
	`
	rows, err := r.db.Query(ctx, query, roomTypeID)
	if err != nil {
		return nil, fmt.Errorf("list price rules: %w", err)
	}
	defer rows.Close()

	var rules []entity.PriceRule
	for rows.Next() {
		var pr entity.PriceRule
		err := rows.Scan(
			&pr.ID, 
			&pr.RoomTypeID, 
			&pr.Start,
			&pr.End,
			&pr.Price, 
			&pr.Priority, 
			&pr.CreatedAt, 
			&pr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rules = append(rules, pr)
	}
	return rules, nil
}

func (r *PriceRepository) GetByID(ctx context.Context, id string) (*entity.PriceRule, error) {
	query := `
		SELECT 
			id, room_type_id, 
			LOWER(validity_range) as start_date, 
			UPPER(validity_range) as end_date, 
			price, priority
		FROM price_rules 
		WHERE id = $1 AND deleted_at IS NULL
	`
	var pr entity.PriceRule
	err := r.db.QueryRow(ctx, query, id).Scan(
		&pr.ID, &pr.RoomTypeID, 
		&pr.Start, &pr.End, 
		&pr.Price, &pr.Priority,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get price rule: %w", err)
	}
	return &pr, nil
}

func (r *PriceRepository) Update(ctx context.Context, id string, req entity.UpdatePriceRuleRequest) error {
	query := `
		UPDATE price_rules 
		SET price = $2, priority = $3, updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`
	cmd, err := r.db.Exec(ctx, query, id, req.Price, req.Priority)
	if err != nil {
		return fmt.Errorf("update price rule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
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
