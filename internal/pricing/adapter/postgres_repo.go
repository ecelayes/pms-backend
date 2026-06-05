package adapter

import (
	"context"
	"errors"
	"fmt"
	"github.com/ecelayes/pms-backend/internal/pricing/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresRatePlanRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRatePlanRepository(db *pgxpool.Pool) *PostgresRatePlanRepository {
	return &PostgresRatePlanRepository{db: db}
}
func (r *PostgresRatePlanRepository) Save(ctx context.Context, rp *domain.RatePlan) error {
	query := `
		INSERT INTO rate_plans (
			id, property_id, unit_type_id, name, description, active,
			meal_plan, cancellation_policy, payment_policy, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET 
		name=$4, description=$5, active=$6, meal_plan=$7, cancellation_policy=$8, payment_policy=$9
	`
	_, err := r.db.Exec(ctx, query,
		rp.ID(), rp.PropertyID(), rp.UnitTypeID(), rp.Name(), rp.Description(), rp.Active(),
		rp.MealPlan(), rp.CancellationPolicy(), rp.PaymentPolicy(), time.Now(),
	)
	return err
}
func (r *PostgresRatePlanRepository) FindByID(ctx context.Context, id string) (*domain.RatePlan, error) {
	query := `
		SELECT property_id, unit_type_id, name, description, active, 
		       meal_plan, cancellation_policy, payment_policy, created_at
		FROM rate_plans WHERE id=$1 AND deleted_at IS NULL
	`
	var propID, name, description string
	var unitTypeID *string
	var active bool
	var mealPlan domain.MealPlan
	var cp domain.CancellationPolicy
	var pp domain.PaymentPolicy
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&propID, &unitTypeID, &name, &description, &active, &mealPlan, &cp, &pp, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return domain.ReconstituteRatePlan(id, propID, unitTypeID, name, description, active, mealPlan, cp, pp, createdAt), nil
}
func (r *PostgresRatePlanRepository) FindByPropertyID(ctx context.Context, propertyID string) ([]*domain.RatePlan, error) {
	query := `
		SELECT id, unit_type_id, name, description, active, 
		       meal_plan, cancellation_policy, payment_policy, created_at
		FROM rate_plans WHERE property_id=$1 AND deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.RatePlan
	for rows.Next() {
		var id, name, description string
		var unitTypeID *string
		var active bool
		var mealPlan domain.MealPlan
		var cp domain.CancellationPolicy
		var pp domain.PaymentPolicy
		var createdAt time.Time
		if err := rows.Scan(&id, &unitTypeID, &name, &description, &active, &mealPlan, &cp, &pp, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, domain.ReconstituteRatePlan(id, propertyID, unitTypeID, name, description, active, mealPlan, cp, pp, createdAt))
	}
	return result, nil
}
func (r *PostgresRatePlanRepository) Delete(ctx context.Context, id string) error {
	var count int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM reservations WHERE rate_plan_id=$1 AND status != 'cancelled'", id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("active reservations depend on it")
	}
	tag, err := r.db.Exec(ctx, "UPDATE rate_plans SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type PostgresPriceRuleRepository struct {
	db *pgxpool.Pool
}

func NewPostgresPriceRuleRepository(db *pgxpool.Pool) *PostgresPriceRuleRepository {
	return &PostgresPriceRuleRepository{db: db}
}
func (r *PostgresPriceRuleRepository) Save(ctx context.Context, pr *domain.PriceRule) error {
	rangeStr := fmt.Sprintf("[%s,%s)",
		pr.DateRange().Start().Format("2006-01-02"),
		pr.DateRange().End().Format("2006-01-02"),
	)
	query := `
		INSERT INTO price_rules (
			id, unit_type_id, validity_range, price_cents, price_currency, price, created_at
		) VALUES ($1, $2, $3::daterange, $4, $5, $7, $6)
		ON CONFLICT (id) DO UPDATE SET 
		validity_range=$3::daterange, price_cents=$4, price=$7
	`
	_, err := r.db.Exec(ctx, query,
		pr.ID(), pr.UnitTypeID(), rangeStr,
		pr.Price().Amount(), pr.Price().Currency(),
		time.Now(),
		float64(pr.Price().Amount())/100.0,
	)
	return err
}
func (r *PostgresPriceRuleRepository) FindOverlapping(ctx context.Context, unitTypeID string, start, end string) ([]*domain.PriceRule, error) {
	rangeStr := fmt.Sprintf("[%s,%s)", start, end)
	query := `
		SELECT id, lower(validity_range), upper(validity_range), price_cents, price_currency, created_at
		FROM price_rules 
		WHERE unit_type_id=$1 AND validity_range && $2::daterange
	`
	rows, err := r.db.Query(ctx, query, unitTypeID, rangeStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.PriceRule
	for rows.Next() {
		var id, currency string
		var startT, endT time.Time
		var priceCents int64
		var createdAt time.Time
		if err := rows.Scan(&id, &startT, &endT, &priceCents, &currency, &createdAt); err != nil {
			return nil, err
		}
		dr, err := vo.NewDateRange(startT, endT)
		if err != nil {
			continue
		}
		money := vo.NewMoney(priceCents, currency)
		result = append(result, domain.ReconstitutePriceRule(id, unitTypeID, dr, money, createdAt))
	}
	return result, nil
}
func (r *PostgresPriceRuleRepository) FindByUnitType(ctx context.Context, unitTypeID string) ([]*domain.PriceRule, error) {
	query := `
		SELECT id, lower(validity_range), upper(validity_range), price_cents, price_currency, created_at
		FROM price_rules WHERE unit_type_id=$1
		ORDER BY lower(validity_range) ASC
	`
	return r.queryRules(ctx, query, unitTypeID)
}
func (r *PostgresPriceRuleRepository) FindByPropertyID(ctx context.Context, propertyID string) ([]*domain.PriceRule, error) {
	query := `
		SELECT pr.id, pr.unit_type_id, lower(pr.validity_range), upper(pr.validity_range), pr.price_cents, pr.price_currency, pr.created_at
		FROM price_rules pr
		JOIN unit_types ut ON pr.unit_type_id = ut.id
		WHERE ut.property_id = $1
		ORDER BY lower(pr.validity_range) ASC
	`
	rows, err := r.db.Query(ctx, query, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.PriceRule
	for rows.Next() {
		var id, unitTypeID, currency string
		var startT, endT time.Time
		var priceCents int64
		var createdAt time.Time
		if err := rows.Scan(&id, &unitTypeID, &startT, &endT, &priceCents, &currency, &createdAt); err != nil {
			return nil, err
		}
		dr, err := vo.NewDateRange(startT, endT)
		if err != nil {
			continue
		}
		money := vo.NewMoney(priceCents, currency)
		result = append(result, domain.ReconstitutePriceRule(id, unitTypeID, dr, money, createdAt))
	}
	return result, nil
}
func (r *PostgresPriceRuleRepository) queryRules(ctx context.Context, query string, args ...interface{}) ([]*domain.PriceRule, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.PriceRule
	for rows.Next() {
		var id, currency string
		var startT, endT time.Time
		var priceCents int64
		var createdAt time.Time
		if err := rows.Scan(&id, &startT, &endT, &priceCents, &currency, &createdAt); err != nil {
			return nil, err
		}
		dr, err := vo.NewDateRange(startT, endT)
		if err != nil {
			continue
		}
		money := vo.NewMoney(priceCents, currency)
		unitTypeID, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("invalid arg")
		}
		result = append(result, domain.ReconstitutePriceRule(id, unitTypeID, dr, money, createdAt))
	}
	return result, nil
}
func (r *PostgresPriceRuleRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "DELETE FROM price_rules WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
