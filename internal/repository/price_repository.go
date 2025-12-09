package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type PriceRepository struct {
	db *pgxpool.Pool
}

func NewPriceRepository(db *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{db: db}
}

func (r *PriceRepository) CreateRule(ctx context.Context, rule entity.PriceRule) error {
	rangeStr := fmt.Sprintf("[%s, %s)", rule.Start.Format("2006-01-02"), rule.End.Format("2006-01-02"))
	query := `
		INSERT INTO price_rules (room_type_id, validity_range, price, priority)
		VALUES ($1, $2::daterange, $3, $4)
	`
	_, err := r.db.Exec(ctx, query, rule.RoomTypeID, rangeStr, rule.Price, rule.Priority)
	if err != nil {
		return fmt.Errorf("insert price rule: %w", err)
	}
	return nil
}

func (r *PriceRepository) Update(ctx context.Context, id string, req entity.UpdatePriceRuleRequest) error {
	query := `UPDATE price_rules SET price = $2, priority = $3 WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id, req.Price, req.Priority)
	if err != nil {
		return fmt.Errorf("update price rule: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("price rule not found")
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
		return fmt.Errorf("price rule not found")
	}
	return nil
}
