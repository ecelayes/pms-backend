package adapter

import (
	"context"
	"errors"
	"encoding/json"
	"fmt"
	"github.com/ecelayes/pms-backend/internal/booking/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresReservationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresReservationRepository(db *pgxpool.Pool) *PostgresReservationRepository {
	return &PostgresReservationRepository{db: db}
}

type ReservationCreatedEvent struct {
	ReservationID string    `json:"reservation_id"`
	PropertyID    string    `json:"property_id"`
	UnitID        string    `json:"unit_id"`
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	OccurredAt    time.Time `json:"occurred_at"`
}
type txKey struct{}

func (r *PostgresReservationRepository) getExecutor(ctx context.Context) (pgx.Tx, bool, error) {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx, true, nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, false, err
	}
	return tx, false, nil
}
func (r *PostgresReservationRepository) Save(ctx context.Context, res *domain.Reservation) error {
	tx, isNested, err := r.getExecutor(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	if !isNested {
		defer func() { _ = tx.Rollback(ctx) }()
	}
	rangeStr := fmt.Sprintf("[%s,%s)",
		res.DateRange().Start().Format("2006-01-02"),
		res.DateRange().End().Format("2006-01-02"),
	)
	var unitID *string
	if res.UnitID() != "" {
		uid := res.UnitID()
		unitID = &uid
	}
	query := `
		INSERT INTO reservations (
			id, property_id, unit_type_id, unit_id, guest_id, stay_range, 
			price_cents, price_currency, status, total_price,
			guest_email, reservation_code, created_at, rate_plan_id
		)
		VALUES ($1, $2, $3, $4, $5, $6::daterange, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (id) DO UPDATE SET status=$9
	`
	_, err = tx.Exec(ctx, query,
		res.ID(),
		res.PropertyID(),
		res.UnitTypeID(),
		unitID,
		res.GuestID(),
		rangeStr,
		res.Price().Amount(),
		res.Price().Currency(),
		res.Status(),
		float64(res.Price().Amount())/100.0,
		res.GuestEmail(),
		res.ReservationCode(),
		time.Now(),
		res.RatePlanID(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert reservation: %w", err)
	}
	eventPayload := ReservationCreatedEvent{
		ReservationID: res.ID(),
		PropertyID:    res.PropertyID(),
		UnitID:        res.UnitID(),
		Start:         res.DateRange().Start(),
		End:           res.DateRange().End(),
		OccurredAt:    time.Now(),
	}
	payloadBytes, err := json.Marshal(eventPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	outboxQuery := `
		INSERT INTO outbox (id, aggregate_id, type, payload, occurred_on)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, outboxQuery,
		res.ID(),
		res.ID(),
		"ReservationCreated",
		payloadBytes,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}
	if !isNested {
		return tx.Commit(ctx)
	}
	return nil
}
func (r *PostgresReservationRepository) FindByID(ctx context.Context, id string) (*domain.Reservation, error) {
	query := `
		SELECT property_id, unit_type_id, unit_id, guest_id, lower(stay_range), upper(stay_range), 
		       price_cents, price_currency, status, guest_email, reservation_code, created_at, rate_plan_id
		FROM reservations WHERE id = $1
	`
	row := r.db.QueryRow(ctx, query, id)
	var propertyID, unitTypeID, currency, statusStr, guestEmail, resCode, ratePlanID string
	var unitID, guestID *string
	var priceCents int64
	var start, end, createdAt time.Time
	err := row.Scan(&propertyID, &unitTypeID, &unitID, &guestID, &start, &end, &priceCents, &currency, &statusStr, &guestEmail, &resCode, &createdAt, &ratePlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find reservation: %w", err)
	}
	dr, err := vo.NewDateRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("data corruption: invalid date range: %w", err)
	}
	money := vo.NewMoney(priceCents, currency)
	uid := ""
	if unitID != nil {
		uid = *unitID
	}
	gid := ""
	if guestID != nil {
		gid = *guestID
	}
	return domain.Reconstitute(
		id,
		propertyID,
		unitTypeID,
		ratePlanID,
		uid,
		gid,
		dr,
		money,
		domain.ReservationStatus(statusStr),
		guestEmail,
		resCode,
		createdAt,
	), nil
}
func (r *PostgresReservationRepository) CountOverlapping(ctx context.Context, unitTypeID string, start, end time.Time) (int, error) {
	rangeStr := fmt.Sprintf("[%s,%s)", start.Format("2006-01-02"), end.Format("2006-01-02"))
	query := `
		SELECT COUNT(*) 
		FROM reservations 
		WHERE unit_type_id = $1 
		AND status = 'confirmed' 
		AND stay_range && $2::daterange
	`
	var count int
	err := r.db.QueryRow(ctx, query, unitTypeID, rangeStr).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (r *PostgresReservationRepository) Update(ctx context.Context, res *domain.Reservation) error {
	rangeStr := fmt.Sprintf("[%s,%s)",
		res.DateRange().Start().Format("2006-01-02"),
		res.DateRange().End().Format("2006-01-02"),
	)
	query := `
		UPDATE reservations 
		SET status=$1, stay_range=$2::daterange, price_cents=$3, price_currency=$4
		WHERE id=$5
	`
	_, err := r.db.Exec(ctx, query,
		res.Status(),
		rangeStr,
		res.Price().Amount(),
		res.Price().Currency(),
		res.ID(),
	)
	return err
}
func (r *PostgresReservationRepository) GetByCode(ctx context.Context, code string) (*domain.Reservation, error) {
	query := `
		SELECT id, property_id, unit_type_id, unit_id, guest_id, lower(stay_range), upper(stay_range), 
		       price_cents, price_currency, status, guest_email, reservation_code, created_at, rate_plan_id
		FROM reservations WHERE reservation_code = $1
	`
	row := r.db.QueryRow(ctx, query, code)
	var id, propertyID, unitTypeID, currency, statusStr, guestEmail, resCode, ratePlanID string
	var unitID, guestID *string
	var priceCents int64
	var start, end, createdAt time.Time
	err := row.Scan(&id, &propertyID, &unitTypeID, &unitID, &guestID, &start, &end, &priceCents, &currency, &statusStr, &guestEmail, &resCode, &createdAt, &ratePlanID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find reservation by code: %w", err)
	}
	dr, err := vo.NewDateRange(start, end)
	if err != nil {
		return nil, fmt.Errorf("data corruption: invalid date range: %w", err)
	}
	money := vo.NewMoney(priceCents, currency)
	uid := ""
	if unitID != nil {
		uid = *unitID
	}
	gid := ""
	if guestID != nil {
		gid = *guestID
	}
	return domain.Reconstitute(
		id,
		propertyID,
		unitTypeID,
		ratePlanID,
		uid,
		gid,
		dr,
		money,
		domain.ReservationStatus(statusStr),
		guestEmail,
		resCode,
		createdAt,
	), nil
}
func (r *PostgresReservationRepository) LockUnitType(ctx context.Context, unitTypeID string) error {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	if !ok {
		return fmt.Errorf("LockUnitType must be run within a transaction")
	}
	_, err := tx.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", unitTypeID)
	return err
}
func (r *PostgresReservationRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	ctxWithTx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctxWithTx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
