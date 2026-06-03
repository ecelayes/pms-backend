package adapter

import (
	"context"
	"github.com/ecelayes/pms-backend/internal/catalog/domain"
	"github.com/ecelayes/pms-backend/internal/shared/vo"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type PostgresCatalogRepository struct {
	db *pgxpool.Pool
}

func NewPostgresCatalogRepository(db *pgxpool.Pool) *PostgresCatalogRepository {
	return &PostgresCatalogRepository{db: db}
}

type PostgresAmenityRepository struct {
	db *pgxpool.Pool
}

func NewPostgresAmenityRepository(db *pgxpool.Pool) *PostgresAmenityRepository {
	return &PostgresAmenityRepository{db: db}
}
func (r *PostgresAmenityRepository) Save(ctx context.Context, a *domain.Amenity) error {
	query := `
		INSERT INTO amenities (id, name, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET name=$2, description=$3, icon=$4, updated_at=NOW()
	`
	_, err := r.db.Exec(ctx, query, a.ID(), a.Name(), a.Description(), a.Icon(), time.Now())
	return err
}
func (r *PostgresAmenityRepository) FindAll(ctx context.Context, limit, offset int) ([]*domain.Amenity, int64, error) {
	query := `
		SELECT id, name, description, icon, COUNT(*) OVER() 
		FROM amenities 
		WHERE deleted_at IS NULL 
		ORDER BY name ASC 
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.Amenity
	var totalCount int64
	for rows.Next() {
		var id, name, desc, icon string
		if err := rows.Scan(&id, &name, &desc, &icon, &totalCount); err != nil {
			return nil, 0, err
		}
		list = append(list, domain.ReconstituteAmenity(id, name, desc, icon))
	}
	return list, totalCount, nil
}
func (r *PostgresAmenityRepository) FindByID(ctx context.Context, id string) (*domain.Amenity, error) {
	query := `SELECT id, name, description, icon FROM amenities WHERE id = $1 AND deleted_at IS NULL`
	var vid, name, desc, icon string
	err := r.db.QueryRow(ctx, query, id).Scan(&vid, &name, &desc, &icon)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return domain.ReconstituteAmenity(vid, name, desc, icon), nil
}
func (r *PostgresAmenityRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE amenities SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *PostgresGuestServiceRepository) DeleteGuestService(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE guest_services SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type PostgresGuestServiceRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGuestServiceRepository(db *pgxpool.Pool) *PostgresGuestServiceRepository {
	return &PostgresGuestServiceRepository{db: db}
}
func (r *PostgresGuestServiceRepository) SaveGuestService(ctx context.Context, gs *domain.GuestService) error {
	query := `
		INSERT INTO guest_services (id, name, description, icon, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET name=$2, description=$3, icon=$4, updated_at=NOW()
	`
	_, err := r.db.Exec(ctx, query, gs.ID(), gs.Name(), gs.Description(), gs.Icon(), time.Now())
	return err
}
func (r *PostgresGuestServiceRepository) FindAllGuestServices(ctx context.Context, limit, offset int) ([]*domain.GuestService, int64, error) {
	query := `
		SELECT id, name, description, icon, COUNT(*) OVER() 
		FROM guest_services 
		WHERE deleted_at IS NULL 
		ORDER BY name ASC 
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*domain.GuestService
	var totalCount int64
	for rows.Next() {
		var id, name, desc, icon string
		if err := rows.Scan(&id, &name, &desc, &icon, &totalCount); err != nil {
			return nil, 0, err
		}
		list = append(list, domain.ReconstituteGuestService(id, name, desc, icon))
	}
	return list, totalCount, nil
}
func (r *PostgresGuestServiceRepository) FindGuestServiceByID(ctx context.Context, id string) (*domain.GuestService, error) {
	query := `SELECT id, name, description, icon FROM guest_services WHERE id = $1 AND deleted_at IS NULL`
	var vid, name, desc, icon string
	err := r.db.QueryRow(ctx, query, id).Scan(&vid, &name, &desc, &icon)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return domain.ReconstituteGuestService(vid, name, desc, icon), nil
}
func (r *PostgresCatalogRepository) SaveProperty(ctx context.Context, p *domain.Property) error {
	query := `
		INSERT INTO properties (id, organization_id, name, code, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET name=$3, code=$4, type=$5
	`
	_, err := r.db.Exec(ctx, query, p.ID(), p.OrganizationID(), p.Name(), p.Code(), p.Type(), time.Now())
	return err
}
func (r *PostgresCatalogRepository) FindPropertyByID(ctx context.Context, id string) (*domain.Property, error) {
	var orgID, name, code, pTypeStr string
	var createdAt time.Time
	query := `SELECT organization_id, name, code, type, created_at FROM properties WHERE id=$1`
	err := r.db.QueryRow(ctx, query, id).Scan(&orgID, &name, &code, &pTypeStr, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return domain.ReconstituteProperty(id, orgID, name, code, pTypeStr, createdAt), nil
}
func (r *PostgresCatalogRepository) FindPropertiesByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*domain.Property, int64, error) {
	query := `
		SELECT id, organization_id, name, code, type, created_at, COUNT(*) OVER() 
		FROM properties 
		WHERE organization_id=$1 AND deleted_at IS NULL
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*domain.Property
	var totalCount int64
	for rows.Next() {
		var id, oID, name, code, pTypeStr string
		var createdAt time.Time
		if err := rows.Scan(&id, &oID, &name, &code, &pTypeStr, &createdAt, &totalCount); err != nil {
			return nil, 0, err
		}
		result = append(result, domain.ReconstituteProperty(id, oID, name, code, pTypeStr, createdAt))
	}
	return result, totalCount, nil
}
func (r *PostgresCatalogRepository) DeleteProperty(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE properties SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *PostgresCatalogRepository) SaveUnitType(ctx context.Context, ut *domain.UnitType) error {
	query := `
		INSERT INTO unit_types (
			id, property_id, name, code, total_quantity, 
			base_price_cents, base_price_currency, 
			max_occupancy, max_adults, max_children, amenities, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (id) DO UPDATE SET 
		name=$3, code=$4, total_quantity=$5, 
		base_price_cents=$6, base_price_currency=$7,
		max_occupancy=$8, max_adults=$9, max_children=$10, amenities=$11
	`
	_, err := r.db.Exec(ctx, query,
		ut.ID(), ut.PropertyID(), ut.Name(), ut.Code(), ut.TotalQuantity(),
		ut.BasePrice().Amount(), ut.BasePrice().Currency(),
		ut.MaxOccupancy(), ut.MaxAdults(), ut.MaxChildren(), ut.Amenities(), time.Now(),
	)
	return err
}
func (r *PostgresCatalogRepository) FindUnitTypeByID(ctx context.Context, id string) (*domain.UnitType, error) {
	query := `
		SELECT property_id, name, code, total_quantity, base_price_cents, base_price_currency, 
		       max_occupancy, max_adults, max_children, amenities, created_at
		FROM unit_types WHERE id=$1
	`
	var propID, name, code, currency string
	var totalQty, maxOcc, maxAd, maxCh int
	var priceCents int64
	var amenities []string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(
		&propID, &name, &code, &totalQty, &priceCents, &currency,
		&maxOcc, &maxAd, &maxCh, &amenities, &createdAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	money := vo.NewMoney(priceCents, currency)
	return domain.ReconstituteUnitType(id, propID, name, code, totalQty, money, maxOcc, maxAd, maxCh, amenities, createdAt), nil
}
func (r *PostgresCatalogRepository) DeleteUnitType(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE unit_types SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *PostgresCatalogRepository) FindUnitTypesByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.UnitType, int64, error) {
	query := `
		SELECT id, name, code, total_quantity, base_price_cents, base_price_currency, 
		       max_occupancy, max_adults, max_children, amenities, created_at, COUNT(*) OVER()
		FROM unit_types 
		WHERE ($1 = '' OR property_id::text = $1) AND deleted_at IS NULL
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, propertyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*domain.UnitType
	var totalCount int64
	for rows.Next() {
		var id, name, code, currency string
		var totalQty, maxOcc, maxAd, maxCh int
		var priceCents int64
		var amenities []string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &code, &totalQty, &priceCents, &currency, &maxOcc, &maxAd, &maxCh, &amenities, &createdAt, &totalCount); err != nil {
			return nil, 0, err
		}
		money := vo.NewMoney(priceCents, currency)
		result = append(result, domain.ReconstituteUnitType(id, propertyID, name, code, totalQty, money, maxOcc, maxAd, maxCh, amenities, createdAt))
	}
	return result, totalCount, nil
}
func (r *PostgresCatalogRepository) SaveUnit(ctx context.Context, u *domain.Unit) error {
	query := `
		INSERT INTO units (id, property_id, unit_type_id, name, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET name=$4, status=$5
	`
	_, err := r.db.Exec(ctx, query, u.ID(), u.PropertyID(), u.UnitTypeID(), u.Name(), u.Status(), time.Now())
	return err
}
func (r *PostgresCatalogRepository) FindUnitByID(ctx context.Context, id string) (*domain.Unit, error) {
	query := `SELECT property_id, unit_type_id, name, status, created_at FROM units WHERE id=$1`
	var propID, typeID, name, status string
	var createdAt time.Time
	err := r.db.QueryRow(ctx, query, id).Scan(&propID, &typeID, &name, &status, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return domain.ReconstituteUnit(id, propID, typeID, name, status, createdAt), nil
}
func (r *PostgresCatalogRepository) DeleteUnit(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, "UPDATE units SET deleted_at=NOW() WHERE id=$1", id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
func (r *PostgresCatalogRepository) FindUnitsByPropertyID(ctx context.Context, propertyID string, limit, offset int) ([]*domain.Unit, int64, error) {
	query := `
		SELECT id, unit_type_id, name, status, created_at, COUNT(*) OVER() 
		FROM units 
		WHERE property_id=$1 AND deleted_at IS NULL
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, propertyID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var result []*domain.Unit
	var totalCount int64
	for rows.Next() {
		var id, typeID, name, status string
		var createdAt time.Time
		if err := rows.Scan(&id, &typeID, &name, &status, &createdAt, &totalCount); err != nil {
			return nil, 0, err
		}
		result = append(result, domain.ReconstituteUnit(id, propertyID, typeID, name, status, createdAt))
	}
	return result, totalCount, nil
}
