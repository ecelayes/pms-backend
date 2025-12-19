package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UnitTypeRepository struct {
	db *pgxpool.Pool
}

func NewUnitTypeRepository(db *pgxpool.Pool) *UnitTypeRepository {
	return &UnitTypeRepository{db: db}
}

func (r *UnitTypeRepository) Create(ctx context.Context, ut entity.UnitType) (string, error) {
	query := `
		INSERT INTO unit_types (
			property_id, name, code, total_quantity, base_price, 
			max_occupancy, max_adults, max_children, amenities,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, 
		ut.PropertyID, ut.Name, ut.Code, ut.TotalQuantity, ut.BasePrice,
		ut.MaxOccupancy, ut.MaxAdults, ut.MaxChildren, ut.Amenities,
	).Scan(&id)
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", entity.ErrConflict
		}
		return "", fmt.Errorf("create unit type: %w", err)
	}
	return id, nil
}

func (r *UnitTypeRepository) GetAll(ctx context.Context) ([]entity.UnitType, error) {
	query := `
		SELECT id, property_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM unit_types
		WHERE deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all unit types: %w", err)
	}
	defer rows.Close()

	var unitTypes []entity.UnitType
	for rows.Next() {
		var ut entity.UnitType
		err := rows.Scan(
			&ut.ID, &ut.PropertyID, &ut.Name, &ut.Code, &ut.TotalQuantity, &ut.BasePrice,
			&ut.MaxOccupancy, &ut.MaxAdults, &ut.MaxChildren, &ut.Amenities,
			&ut.CreatedAt, &ut.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		unitTypes = append(unitTypes, ut)
	}
	return unitTypes, nil
}

func (r *UnitTypeRepository) GetByID(ctx context.Context, id string) (*entity.UnitType, error) {
	query := `
		SELECT id, property_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM unit_types
		WHERE id = $1 AND deleted_at IS NULL
	`
	var ut entity.UnitType
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ut.ID, &ut.PropertyID, &ut.Name, &ut.Code, &ut.TotalQuantity, &ut.BasePrice,
		&ut.MaxOccupancy, &ut.MaxAdults, &ut.MaxChildren, &ut.Amenities,
		&ut.CreatedAt, &ut.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get unit type: %w", err)
	}
	return &ut, nil
}

func (r *UnitTypeRepository) GetByIDLocked(ctx context.Context, tx pgx.Tx, id string) (*entity.UnitType, error) {
	query := `
		SELECT id, property_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM unit_types
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE
	`
	var ut entity.UnitType
	err := tx.QueryRow(ctx, query, id).Scan(
		&ut.ID, &ut.PropertyID, &ut.Name, &ut.Code, &ut.TotalQuantity, &ut.BasePrice,
		&ut.MaxOccupancy, &ut.MaxAdults, &ut.MaxChildren, &ut.Amenities,
		&ut.CreatedAt, &ut.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get unit type locked: %w", err)
	}
	return &ut, nil
}

func (r *UnitTypeRepository) ListByProperty(ctx context.Context, propertyID string, pagination entity.PaginationRequest) ([]entity.UnitType, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM unit_types
		WHERE property_id = $1 AND deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, propertyID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count unit types: %w", err)
	}

	query := `
		SELECT id, property_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM unit_types
		WHERE property_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, propertyID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list unit types: %w", err)
	}
	defer rows.Close()

	var unitTypes []entity.UnitType
	for rows.Next() {
		var ut entity.UnitType
		err := rows.Scan(
			&ut.ID, &ut.PropertyID, &ut.Name, &ut.Code, &ut.TotalQuantity, &ut.BasePrice,
			&ut.MaxOccupancy, &ut.MaxAdults, &ut.MaxChildren, &ut.Amenities,
			&ut.CreatedAt, &ut.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		unitTypes = append(unitTypes, ut)
	}
	return unitTypes, total, nil
}

func (r *UnitTypeRepository) Update(ctx context.Context, id string, req entity.UpdateUnitTypeRequest) error {
	query := `UPDATE unit_types SET updated_at = NOW()`
	args := []interface{}{}
	argID := 1

	addSet := func(column string, value interface{}) {
		query += fmt.Sprintf(", %s = $%d", column, argID)
		args = append(args, value)
		argID++
	}

	if req.Name != "" {
		addSet("name", req.Name)
	}
	if req.Code != "" {
		addSet("code", req.Code)
	}

	if req.TotalQuantity != nil {
		if *req.TotalQuantity < 0 {
			return fmt.Errorf("total quantity cannot be negative")
		}
		addSet("total_quantity", *req.TotalQuantity)
	}

	if req.MaxOccupancy != nil {
		if *req.MaxOccupancy <= 0 {
			return fmt.Errorf("max occupancy must be positive")
		}
		addSet("max_occupancy", *req.MaxOccupancy)
	}

	if req.MaxAdults != nil {
		if *req.MaxAdults <= 0 {
			return fmt.Errorf("max adults must be positive")
		}
		addSet("max_adults", *req.MaxAdults)
	}

	if req.MaxChildren != nil {
		if *req.MaxChildren < 0 {
			return fmt.Errorf("max children cannot be negative")
		}
		addSet("max_children", *req.MaxChildren)
	}

	if req.BasePrice != nil {
		if *req.BasePrice < 0 {
			return fmt.Errorf("base price cannot be negative")
		}
		addSet("base_price", *req.BasePrice)
	}

	if req.Amenities != nil {
		addSet("amenities", req.Amenities)
	}

	query += fmt.Sprintf(" WHERE id = $%d AND deleted_at IS NULL", argID)
	args = append(args, id)

	cmd, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update unit type: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}

	return nil
}

func (r *UnitTypeRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE unit_types SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete unit type: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *UnitTypeRepository) GetCodesForGeneration(ctx context.Context, unitTypeID string) (string, string, error) {
	query := `
		SELECT p.code, ut.code
		FROM unit_types ut
		JOIN properties p ON ut.property_id = p.id
		WHERE ut.id = $1 AND ut.deleted_at IS NULL AND p.deleted_at IS NULL
	`
	var propertyCode, unitTypeCode string
	err := r.db.QueryRow(ctx, query, unitTypeID).Scan(&propertyCode, &unitTypeCode)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch codes: %w", err)
	}
	return propertyCode, unitTypeCode, nil
}

func (r *UnitTypeRepository) CountReservations(ctx context.Context, db DBTX, unitTypeID string, start, end time.Time) (int, error) {
	var querier DBTX = db
	if querier == nil {
		querier = r.db
	}
	query := `
		SELECT COUNT(*) FROM reservations 
		WHERE unit_type_id = $1 
		AND status = 'confirmed' 
		AND deleted_at IS NULL
		AND stay_range && daterange($2::date, $3::date)
	`
	var count int
	err := querier.QueryRow(ctx, query, unitTypeID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count reservations: %w", err)
	}
	return count, nil
}

func (r *UnitTypeRepository) GetDailyPrices(ctx context.Context, unitTypeID string, start, end time.Time) ([]entity.DailyRate, error) {
	query := `
		WITH booking_days AS (
			SELECT generate_series($2::date, $3::date - interval '1 day', '1 day')::date AS day
		),
		ranked_prices AS (
			SELECT bd.day, pr.price, ROW_NUMBER() OVER (PARTITION BY bd.day ORDER BY pr.priority DESC) as rn
			FROM booking_days bd
			JOIN price_rules pr ON pr.unit_type_id = $1 AND pr.validity_range @> bd.day
			WHERE pr.deleted_at IS NULL
		)
		SELECT day::text, price FROM ranked_prices WHERE rn = 1 ORDER BY day;
	`
	rows, err := r.db.Query(ctx, query, unitTypeID, start, end)
	if err != nil {
		return nil, fmt.Errorf("calc prices: %w", err)
	}
	defer rows.Close()

	var rates []entity.DailyRate
	for rows.Next() {
		var rate entity.DailyRate
		if err := rows.Scan(&rate.Date, &rate.Price); err != nil {
			return nil, err
		}
		rates = append(rates, rate)
	}
	return rates, nil
}
