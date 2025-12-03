package repository

import (
	"context"
	"fmt"
	"time"

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
		INSERT INTO room_types (hotel_id, name, code, total_quantity)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	var id string
	err := r.db.QueryRow(ctx, query, rt.HotelID, rt.Name, rt.Code, rt.TotalQuantity).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("create room type: %w", err)
	}
	return id, nil
}

func (r *RoomRepository) GetAllRoomTypes(ctx context.Context) ([]entity.RoomType, error) {
	query := `SELECT id, hotel_id, name, code, total_quantity, created_at, updated_at FROM room_types WHERE deleted_at IS NULL`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query room types: %w", err)
	}
	defer rows.Close()

	var result []entity.RoomType
	for rows.Next() {
		var rt entity.RoomType
		if err := rows.Scan(&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.CreatedAt, &rt.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *RoomRepository) GetRoomTypeByID(ctx context.Context, id string) (entity.RoomType, error) {
	query := `SELECT id, hotel_id, name, code, total_quantity, created_at, updated_at FROM room_types WHERE id = $1 AND deleted_at IS NULL`
	var rt entity.RoomType
	err := r.db.QueryRow(ctx, query, id).Scan(&rt.ID, &rt.HotelID, &rt.Name, &rt.Code, &rt.TotalQuantity, &rt.CreatedAt, &rt.UpdatedAt)
	if err != nil {
		return entity.RoomType{}, fmt.Errorf("failed to get room type: %w", err)
	}
	return rt, nil
}

func (r *RoomRepository) Update(ctx context.Context, id string, req entity.UpdateRoomTypeRequest) error {
	query := `UPDATE room_types SET name = $2, code = $3, total_quantity = $4 WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id, req.Name, req.Code, req.TotalQuantity)
	if err != nil {
		return fmt.Errorf("update room type: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return fmt.Errorf("room type not found")
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
		return fmt.Errorf("room type not found")
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
			WHERE pr.deleted_at IS NULL -- Filtro Soft Delete
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
