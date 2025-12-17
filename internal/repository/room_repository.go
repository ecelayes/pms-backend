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

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoomType(ctx context.Context, rt entity.RoomType) (string, error) {
	query := `
		INSERT INTO room_types (
			hotel_id, name, code, total_quantity, base_price, 
			max_occupancy, max_adults, max_children, amenities
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, 
		rt.HotelID, rt.Name, rt.Code, rt.TotalQuantity, rt.BasePrice,
		rt.MaxOccupancy, rt.MaxAdults, rt.MaxChildren, rt.Amenities,
	).Scan(&id)
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", entity.ErrConflict
		}
		return "", fmt.Errorf("create room type: %w", err)
	}
	return id, nil
}

func (r *RoomRepository) GetAll(ctx context.Context) ([]entity.RoomType, error) {
	query := `
		SELECT id, hotel_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM room_types
		WHERE deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get all room types: %w", err)
	}
	defer rows.Close()

	var rooms []entity.RoomType
	for rows.Next() {
		var rt entity.RoomType
		err := rows.Scan(
			&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.BasePrice,
			&rt.MaxOccupancy, &rt.MaxAdults, &rt.MaxChildren, &rt.Amenities,
			&rt.CreatedAt, &rt.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rooms = append(rooms, rt)
	}
	return rooms, nil
}

func (r *RoomRepository) GetByID(ctx context.Context, id string) (*entity.RoomType, error) {
	query := `
		SELECT id, hotel_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM room_types
		WHERE id = $1 AND deleted_at IS NULL
	`
	var rt entity.RoomType
	err := r.db.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.BasePrice,
		&rt.MaxOccupancy, &rt.MaxAdults, &rt.MaxChildren, &rt.Amenities,
		&rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get room type: %w", err)
	}
	return &rt, nil
}

func (r *RoomRepository) GetByIDLocked(ctx context.Context, tx pgx.Tx, id string) (*entity.RoomType, error) {
	query := `
		SELECT id, hotel_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM room_types
		WHERE id = $1 AND deleted_at IS NULL
		FOR UPDATE -- <--- LA CLAVE MÁGICA
	`
	var rt entity.RoomType
	err := tx.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.BasePrice,
		&rt.MaxOccupancy, &rt.MaxAdults, &rt.MaxChildren, &rt.Amenities,
		&rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get room type locked: %w", err)
	}
	return &rt, nil
}

func (r *RoomRepository) ListByHotel(ctx context.Context, hotelID string, pagination entity.PaginationRequest) ([]entity.RoomType, int64, error) {
	countQuery := `
		SELECT COUNT(*)
		FROM room_types
		WHERE hotel_id = $1 AND deleted_at IS NULL
	`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, hotelID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count room types: %w", err)
	}

	query := `
		SELECT id, hotel_id, name, code, total_quantity, base_price,
		       max_occupancy, max_adults, max_children, amenities, 
		       created_at, updated_at
		FROM room_types
		WHERE hotel_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	offset := (pagination.Page - 1) * pagination.Limit
	
	rows, err := r.db.Query(ctx, query, hotelID, pagination.Limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list room types: %w", err)
	}
	defer rows.Close()

	var rooms []entity.RoomType
	for rows.Next() {
		var rt entity.RoomType
		err := rows.Scan(
			&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.BasePrice,
			&rt.MaxOccupancy, &rt.MaxAdults, &rt.MaxChildren, &rt.Amenities,
			&rt.CreatedAt, &rt.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		rooms = append(rooms, rt)
	}
	return rooms, total, nil
}

func (r *RoomRepository) Update(ctx context.Context, id string, req entity.UpdateRoomTypeRequest) error {
	query := `UPDATE room_types SET updated_at = NOW()`
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
		return fmt.Errorf("update room type: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}

	return nil
}

func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE room_types SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete room type: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *RoomRepository) GetCodesForGeneration(ctx context.Context, roomTypeID string) (string, string, error) {
	query := `
		SELECT h.code, rt.code
		FROM room_types rt
		JOIN hotels h ON rt.hotel_id = h.id
		WHERE rt.id = $1 AND rt.deleted_at IS NULL AND h.deleted_at IS NULL
	`
	var hotelCode, roomCode string
	err := r.db.QueryRow(ctx, query, roomTypeID).Scan(&hotelCode, &roomCode)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch codes: %w", err)
	}
	return hotelCode, roomCode, nil
}

func (r *RoomRepository) CountReservations(ctx context.Context, db DBTX, roomTypeID string, start, end time.Time) (int, error) {
	var querier DBTX = db
	if querier == nil {
		querier = r.db
	}
	query := `
		SELECT COUNT(*) FROM reservations 
		WHERE room_type_id = $1 
		AND status = 'confirmed' 
		AND deleted_at IS NULL
		AND stay_range && daterange($2::date, $3::date)
	`
	var count int
	err := querier.QueryRow(ctx, query, roomTypeID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count reservations: %w", err)
	}
	return count, nil
}

func (r *RoomRepository) GetDailyPrices(ctx context.Context, roomTypeID string, start, end time.Time) ([]entity.DailyRate, error) {
	query := `
		WITH booking_days AS (
			SELECT generate_series($2::date, $3::date - interval '1 day', '1 day')::date AS day
		),
		ranked_prices AS (
			SELECT bd.day, pr.price, ROW_NUMBER() OVER (PARTITION BY bd.day ORDER BY pr.priority DESC) as rn
			FROM booking_days bd
			JOIN price_rules pr ON pr.room_type_id = $1 AND pr.validity_range @> bd.day
			WHERE pr.deleted_at IS NULL
		)
		SELECT day::text, price FROM ranked_prices WHERE rn = 1 ORDER BY day;
	`
	rows, err := r.db.Query(ctx, query, roomTypeID, start, end)
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
