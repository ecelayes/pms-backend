package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, password, salt, role string) error {
	query := `INSERT INTO users (email, password, salt, role) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, query, email, password, salt, role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return entity.ErrEmailAlreadyExists
			}
		}
		return fmt.Errorf("db error: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	query := `SELECT id, email, password, salt, role FROM users WHERE email=$1`
	err := r.db.QueryRow(ctx, query, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Salt, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("user lookup failed: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetSaltByID(ctx context.Context, userID string) (string, error) {
	var salt string
	query := `SELECT salt FROM users WHERE id=$1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&salt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("failed to fetch salt: %w", err)
	}
	return salt, nil
}
