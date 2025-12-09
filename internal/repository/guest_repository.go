package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type GuestRepository struct {
	db *pgxpool.Pool
}

func NewGuestRepository(db *pgxpool.Pool) *GuestRepository {
	return &GuestRepository{db: db}
}

func (r *GuestRepository) Create(ctx context.Context, tx pgx.Tx, g entity.Guest) (string, error) {
	query := `
		INSERT INTO guests (id, email, first_name, last_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`
	var id string
	var err error
	
	if tx != nil {
		err = tx.QueryRow(ctx, query, g.ID, g.Email, g.FirstName, g.LastName, g.Phone).Scan(&id)
	} else {
		err = r.db.QueryRow(ctx, query, g.ID, g.Email, g.FirstName, g.LastName, g.Phone).Scan(&id)
	}

	if err != nil {
		return "", fmt.Errorf("create guest: %w", err)
	}
	return id, nil
}

func (r *GuestRepository) GetByEmail(ctx context.Context, email string) (*entity.Guest, error) {
	query := `SELECT id, email, first_name, last_name, phone FROM guests WHERE email = $1 AND deleted_at IS NULL`
	var g entity.Guest
	err := r.db.QueryRow(ctx, query, email).Scan(&g.ID, &g.Email, &g.FirstName, &g.LastName, &g.Phone)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get guest by email: %w", err)
	}
	return &g, nil
}

func (r *GuestRepository) Update(ctx context.Context, tx pgx.Tx, g entity.Guest) error {
	query := `
		UPDATE guests 
		SET first_name = $2, last_name = $3, phone = $4, updated_at = NOW()
		WHERE email = $1 AND deleted_at IS NULL
	`
	var err error
	
	if tx != nil {
		_, err = tx.Exec(ctx, query, g.Email, g.FirstName, g.LastName, g.Phone)
	} else {
		_, err = r.db.Exec(ctx, query, g.Email, g.FirstName, g.LastName, g.Phone)
	}

	if err != nil {
		return fmt.Errorf("update guest: %w", err)
	}
	return nil
}
