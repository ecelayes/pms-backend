package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type RoomRepository struct {
	db *pgxpool.Pool
}

func NewRoomRepository(db *pgxpool.Pool) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) GetAllRoomTypes(ctx context.Context) ([]entity.RoomType, error) {
	rows, err := r.db.Query(ctx, "SELECT id, name, total_quantity FROM room_types")
	if err != nil {
		return nil, fmt.Errorf("query room types: %w", err)
	}
	defer rows.Close()

	var result []entity.RoomType
	for rows.Next() {
		var rt entity.RoomType
		if err := rows.Scan(&rt.ID, &rt.Name, &rt.TotalQuantity); err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *RoomRepository) CountReservations(ctx context.Context, db DBTX, roomTypeID string, start, end time.Time) (int, error) {
	var querier DBTX = db
	if querier == nil {
		querier = r.db
	}

	query := `
		SELECT COUNT(*) 
		FROM reservations 
		WHERE room_type_id = $1 
		AND status = 'confirmed'
		AND stay_range && daterange($2::date, $3::date)
	`
	var count int
	err := querier.QueryRow(ctx, query, roomTypeID, start, end).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count reservations: %w", err)
	}
	return count, nil
}

func (r *RoomRepository) GetRoomTypeByID(ctx context.Context, id string) (entity.RoomType, error) {
	query := `
		SELECT id, name, total_quantity 
		FROM room_types 
		WHERE id = $1
	`

	var rt entity.RoomType
	err := r.db.QueryRow(ctx, query, id).Scan(&rt.ID, &rt.Name, &rt.TotalQuantity)
	if err != nil {
		return entity.RoomType{}, fmt.Errorf("failed to get room type: %w", err)
	}

	return rt, nil
}

func (r *RoomRepository) CreateReservation(ctx context.Context, tx pgx.Tx, ent entity.Reservation) error {
	query := `
		INSERT INTO reservations (id, room_type_id, stay_range, guest_email, total_price, status)
		VALUES ($1, $2, daterange($3::date, $4::date), $5, $6, 'confirmed')
	`
	_, err := tx.Exec(ctx, query, 
		ent.ID, 
		ent.RoomTypeID, 
		ent.Start, 
		ent.End, 
		ent.GuestEmail, 
		ent.TotalPrice,
	)
	if err != nil {
		return fmt.Errorf("insert reservation: %w", err)
	}
	return nil
}

func (r *RoomRepository) GetDailyPrices(ctx context.Context, roomTypeID string, start, end time.Time) ([]entity.DailyRate, error) {
	query := `
		WITH booking_days AS (
			-- Generate one row per day in the requested range
			SELECT generate_series($2::date, $3::date - interval '1 day', '1 day')::date AS day
		),
		ranked_prices AS (
			-- Find applicable rules for each day and rank by priority
			SELECT 
				bd.day,
				pr.price,
				ROW_NUMBER() OVER (PARTITION BY bd.day ORDER BY pr.priority DESC) as rn
			FROM booking_days bd
			JOIN price_rules pr ON pr.room_type_id = $1 AND pr.validity_range @> bd.day
		)
		-- Select only the highest priority price per day
		SELECT day::text, price 
		FROM ranked_prices 
		WHERE rn = 1 
		ORDER BY day;
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
