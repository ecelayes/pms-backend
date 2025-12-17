package repository

import (
	"context"
	"fmt"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ecelayes/pms-backend/internal/entity"
)

type OrganizationRepository struct {
	db *pgxpool.Pool
}

func NewOrganizationRepository(db *pgxpool.Pool) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, tx pgx.Tx, org entity.Organization) error {
	query := `
		INSERT INTO organizations (id, name, code, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, org.ID, org.Name, org.Code)
	} else {
		_, err = r.db.Exec(ctx, query, org.ID, org.Name, org.Code)
	}
	
	if err != nil {
		return fmt.Errorf("create org: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) GetAll(ctx context.Context, pagination entity.PaginationRequest) ([]entity.Organization, int64, error) {
	countQuery := `SELECT COUNT(*) FROM organizations WHERE deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count organizations: %w", err)
	}

	var query string
	var args []interface{}

	if pagination.Unlimited {
		query = `
			SELECT id, name, code, created_at, updated_at 
			FROM organizations 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
		`
	} else {
		query = `
			SELECT id, name, code, created_at, updated_at 
			FROM organizations 
			WHERE deleted_at IS NULL 
			ORDER BY name ASC
			LIMIT $1 OFFSET $2
		`
		offset := (pagination.Page - 1) * pagination.Limit
		args = append(args, pagination.Limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	var list []entity.Organization
	for rows.Next() {
		var org entity.Organization
		if err := rows.Scan(&org.ID, &org.Name, &org.Code, &org.CreatedAt, &org.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, org)
	}
	return list, total, nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*entity.Organization, error) {
	query := `
		SELECT id, name, code, created_at, updated_at 
		FROM organizations 
		WHERE id = $1 AND deleted_at IS NULL
	`
	var org entity.Organization
	err := r.db.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Code, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entity.ErrRecordNotFound
		}
		return nil, fmt.Errorf("get org: %w", err)
	}
	return &org, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, id string, req entity.UpdateOrganizationRequest) error {
	query := `UPDATE organizations SET name = $2, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id, req.Name)
	if err != nil {
		return fmt.Errorf("update org: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	query := `UPDATE organizations SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmd, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete org: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return entity.ErrRecordNotFound
	}
	return nil
}

func (r *OrganizationRepository) AddMember(ctx context.Context, tx pgx.Tx, member entity.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, query, member.ID, member.OrganizationID, member.UserID, member.Role)
	} else {
		_, err = r.db.Exec(ctx, query, member.ID, member.OrganizationID, member.UserID, member.Role)
	}
	
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}
	return nil
}

func (r *OrganizationRepository) GetUserOrganization(ctx context.Context, userID string) (string, error) {
	var orgID string
	query := `SELECT organization_id FROM organization_members WHERE user_id = $1 LIMIT 1`
	err := r.db.QueryRow(ctx, query, userID).Scan(&orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil 
		}
		return "", fmt.Errorf("fetch user org: %w", err)
	}
	return orgID, nil
}
