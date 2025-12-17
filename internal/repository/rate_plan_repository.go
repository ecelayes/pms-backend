package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type RatePlanRepository struct {
	db *pgxpool.Pool
}

func NewRatePlanRepository(db *pgxpool.Pool) *RatePlanRepository {
	return &RatePlanRepository{db: db}
}

func (r *RatePlanRepository) Create(ctx context.Context, rp entity.RatePlan) error {
	query := `
		INSERT INTO rate_plans (
			id, hotel_id, room_type_id, name, description, 
			meal_plan, cancellation_policy, payment_policy, active,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query,
		rp.ID, rp.HotelID, rp.RoomTypeID, rp.Name, rp.Description,
		rp.MealPlan, rp.CancellationPolicy, rp.PaymentPolicy, rp.Active,
	)
	if err != nil {
		return fmt.Errorf("create rate plan: %w", err)
	}
	return nil
}

func (r *RatePlanRepository) ListByHotel(ctx context.Context, hotelID string, pagination entity.PaginationRequest) ([]entity.RatePlan, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM rate_plans
		WHERE hotel_id = $1 AND deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, hotelID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count rate plans: %w", err)
	}

	query := `
		SELECT id, hotel_id, room_type_id, name, description, 
		       meal_plan, cancellation_policy, payment_policy, active, created_at, updated_at
		FROM rate_plans
		WHERE hotel_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, hotelID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list rate plans: %w", err)
	}
	defer rows.Close()

	var plans []entity.RatePlan
	for rows.Next() {
		var rp entity.RatePlan
		err := rows.Scan(
			&rp.ID, &rp.HotelID, &rp.RoomTypeID, &rp.Name, &rp.Description,
			&rp.MealPlan, &rp.CancellationPolicy, &rp.PaymentPolicy, &rp.Active,
			&rp.CreatedAt, &rp.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		plans = append(plans, rp)
	}
	return plans, total, nil
}

func (r *RatePlanRepository) GetByID(ctx context.Context, id string) (*entity.RatePlan, error) {
	query := `
		SELECT id, hotel_id, room_type_id, name, description, 
		       meal_plan, cancellation_policy, payment_policy, active, created_at, updated_at
		FROM rate_plans
		WHERE id = $1 AND deleted_at IS NULL
	`
	var rp entity.RatePlan
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rp.ID, &rp.HotelID, &rp.RoomTypeID, &rp.Name, &rp.Description,
		&rp.MealPlan, &rp.CancellationPolicy, &rp.PaymentPolicy, &rp.Active,
		&rp.CreatedAt, &rp.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get rate plan: %w", err)
	}
	return &rp, nil
}

func (r *RatePlanRepository) Update(ctx context.Context, id string, req entity.UpdateRatePlanRequest) error {
	query := `UPDATE rate_plans SET updated_at = NOW()`
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
	if req.Description != "" {
		addSet("description", req.Description)
	}
	if req.Active != nil {
		addSet("active", *req.Active)
	}
	
	if req.MealPlan != nil {
		addSet("meal_plan", req.MealPlan)
	}
	if req.CancellationPolicy != nil {
		addSet("cancellation_policy", req.CancellationPolicy)
	}
	if req.PaymentPolicy != nil {
		addSet("payment_policy", req.PaymentPolicy)
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argID)
	args = append(args, id)

	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update rate plan: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *RatePlanRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE rate_plans SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete rate plan: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *RatePlanRepository) GetAll(ctx context.Context) ([]entity.RatePlan, error) {
	query := `
		SELECT id, hotel_id, room_type_id, name, description, 
		       meal_plan, cancellation_policy, payment_policy, active, created_at, updated_at
		FROM rate_plans
		WHERE deleted_at IS NULL AND active = TRUE
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all active rate plans: %w", err)
	}
	defer rows.Close()

	var plans []entity.RatePlan
	for rows.Next() {
		var rp entity.RatePlan
		err := rows.Scan(
			&rp.ID, &rp.HotelID, &rp.RoomTypeID, &rp.Name, &rp.Description,
			&rp.MealPlan, &rp.CancellationPolicy, &rp.PaymentPolicy, &rp.Active,
			&rp.CreatedAt, &rp.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, rp)
	}
	return plans, nil
}
