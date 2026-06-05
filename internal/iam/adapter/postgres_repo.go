package adapter

import (
	"context"
	"errors"
	"github.com/ecelayes/pms-backend/internal/iam/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}
func (r *PostgresUserRepository) Save(ctx context.Context, u *domain.User) error {
	queryExtended := `
		INSERT INTO users (id, email, password, salt, role, first_name, last_name, phone, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET 
		email=$2, password=$3, salt=$4, role=$5, first_name=$6, last_name=$7, phone=$8
	`
	_, err := r.db.Exec(ctx, queryExtended,
		u.ID(), u.Email(), u.Password(), u.Salt(), u.Role(),
		u.FirstName(), u.LastName(), u.Phone(),
		time.Now(),
	)
	return err
}
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT email, password, salt, role, first_name, last_name, phone, created_at 
		FROM users WHERE id=$1 AND deleted_at IS NULL
	`
	var email, password, salt, role, fName, lName, phone string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(&email, &password, &salt, &role, &fName, &lName, &phone, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return domain.ReconstituteUser(id, email, password, salt, role, fName, lName, phone, createdAt), nil
}
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, password, salt, role, first_name, last_name, phone, created_at 
		FROM users WHERE email=$1 AND deleted_at IS NULL
	`
	var id, password, salt, role, fName, lName, phone string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, email).Scan(&id, &password, &salt, &role, &fName, &lName, &phone, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return domain.ReconstituteUser(id, email, password, salt, role, fName, lName, phone, createdAt), nil
}
func (r *PostgresUserRepository) FindAllByOrganization(ctx context.Context, orgID string) ([]*domain.User, error) {
	query := `
		SELECT u.id, u.email, u.password, u.salt, u.role, u.first_name, u.last_name, u.phone, u.created_at
		FROM users u
		JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = $1 AND u.deleted_at IS NULL
	`
	rows, err := r.db.Query(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.User
	for rows.Next() {
		var id, email, password, salt, role, fName, lName, phone string
		var createdAt time.Time
		if err := rows.Scan(&id, &email, &password, &salt, &role, &fName, &lName, &phone, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, domain.ReconstituteUser(id, email, password, salt, role, fName, lName, phone, createdAt))
	}
	return result, nil
}
func (r *PostgresUserRepository) DeleteUser(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE users SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
func (r *PostgresUserRepository) EnsureGuest(ctx context.Context, u *domain.User) error {
	query := `
		INSERT INTO guests (id, email, first_name, last_name, phone, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (email) DO UPDATE SET 
		id=$1, first_name=$3, last_name=$4, phone=$5
	`
	_, err := r.db.Exec(ctx, query, u.ID(), u.Email(), u.FirstName(), u.LastName(), u.Phone(), time.Now())
	return err
}

type PostgresOrganizationRepository struct {
	db *pgxpool.Pool
}

func NewPostgresOrganizationRepository(db *pgxpool.Pool) *PostgresOrganizationRepository {
	return &PostgresOrganizationRepository{db: db}
}
func (r *PostgresOrganizationRepository) Save(ctx context.Context, o *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, code, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET name=$2, code=$3
	`
	_, err := r.db.Exec(ctx, query, o.ID(), o.Name(), o.Code(), time.Now())
	return err
}
func (r *PostgresOrganizationRepository) FindByID(ctx context.Context, id string) (*domain.Organization, error) {
	if err := uuid.Validate(id); err != nil {
		return nil, domain.ErrNotFound
	}
	query := `SELECT name, code, created_at FROM organizations WHERE id=$1 AND deleted_at IS NULL`
	var name, code string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(&name, &code, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return domain.ReconstituteOrganization(id, name, code, createdAt), nil
}
func (r *PostgresOrganizationRepository) AddMember(ctx context.Context, orgID, userID, role string) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW())
		ON CONFLICT (organization_id, user_id) DO UPDATE SET role=$3
	`
	_, err := r.db.Exec(ctx, query, orgID, userID, role)
	return err
}
func (r *PostgresOrganizationRepository) FindMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	query := `SELECT role FROM organization_members WHERE organization_id=$1 AND user_id=$2`
	var role string
	err := r.db.QueryRow(ctx, query, orgID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	return role, nil
}
func (r *PostgresOrganizationRepository) FindByUserID(ctx context.Context, userID string) (*domain.Organization, error) {
	query := `
		SELECT o.id, o.name, o.code, o.created_at
		FROM organizations o
		JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1 AND o.deleted_at IS NULL
		LIMIT 1
	`
	var id, name, code string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, userID).Scan(&id, &name, &code, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return domain.ReconstituteOrganization(id, name, code, createdAt), nil
}
func (r *PostgresOrganizationRepository) FindAll(ctx context.Context) ([]*domain.Organization, error) {
	query := `SELECT id, name, code, created_at FROM organizations WHERE deleted_at IS NULL ORDER BY name ASC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []*domain.Organization
	for rows.Next() {
		var id, name, code string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &code, &createdAt); err != nil {
			return nil, err
		}
		result = append(result, domain.ReconstituteOrganization(id, name, code, createdAt))
	}
	return result, nil
}
func (r *PostgresOrganizationRepository) DeleteOrganization(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, "UPDATE organizations SET deleted_at=NOW() WHERE id=$1", id)
	return err
}
