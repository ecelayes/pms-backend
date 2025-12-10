package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, tx pgx.Tx, id, email, password, salt, role string) error {
	query := `
		INSERT INTO users (id, email, password, salt, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, id, email, password, salt, role)
	} else {
		_, err = r.db.Exec(ctx, query, id, email, password, salt, role)
	}

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetSaltByID(ctx context.Context, userID string) (string, error) {
	var salt string
	query := `SELECT salt FROM users WHERE id=$1 AND deleted_at IS NULL`
	err := r.db.QueryRow(ctx, query, userID).Scan(&salt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", entity.ErrUserNotFound
		}
		return "", fmt.Errorf("failed to fetch salt: %w", err)
	}
	return salt, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	query := `SELECT id, email, password, salt, role FROM users WHERE email=$1 AND deleted_at IS NULL`
	err := r.db.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.Password, &u.Salt, &u.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("user lookup failed: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, email, password, salt, role, created_at, updated_at 
		FROM users 
		WHERE id = $1 AND deleted_at IS NULL
	`
	var u entity.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Email, &u.Password, &u.Salt, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) GetAllByOrganization(ctx context.Context, orgID string) ([]entity.User, error) {
	query := `
		SELECT u.id, u.email, u.created_at, u.updated_at, om.role
		FROM users u
		JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = $1 AND u.deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("list org users: %w", err)
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, userID, orgID string, req entity.UpdateUserRequest) error {
	if req.Email != "" {
		_, err := r.db.Exec(ctx, "UPDATE users SET email = $2, updated_at = NOW() WHERE id = $1", userID, req.Email)
		if err != nil { return err }
	}

	if req.Role != "" {
		query := `UPDATE organization_members SET role = $3, updated_at = NOW() WHERE user_id = $1 AND organization_id = $2`
		cmd, err := r.db.Exec(ctx, query, userID, orgID, req.Role)
		if err != nil { return err }
		if cmd.RowsAffected() == 0 { return entity.ErrRecordNotFound }
	}
	
	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil { return fmt.Errorf("delete user: %w", err) }
	if cmd.RowsAffected() == 0 { return entity.ErrRecordNotFound }
	return nil
}
