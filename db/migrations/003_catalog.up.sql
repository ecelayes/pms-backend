-- Catalog: Unit Types, Amenities, Guest Services, Units

-- Unit Types table (renamed from room_types)
CREATE TABLE unit_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    code VARCHAR(5) NOT NULL,
    name TEXT NOT NULL,
    total_quantity INT NOT NULL CHECK (total_quantity >= 0),
    max_occupancy INT NOT NULL DEFAULT 2,
    max_adults INT NOT NULL DEFAULT 2,
    max_children INT NOT NULL DEFAULT 0,
    amenities TEXT[],
    base_price DECIMAL(10, 2) NOT NULL DEFAULT 0,
    base_price_cents BIGINT DEFAULT 0,
    base_price_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT unique_room_code_per_hotel UNIQUE (property_id, code)
);
CREATE TRIGGER update_unit_types_modtime 
    BEFORE UPDATE ON unit_types 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE INDEX idx_room_types_deleted_at ON unit_types(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_room_types_property_id ON unit_types(property_id);
CREATE INDEX idx_room_types_amenities ON unit_types USING GIN (amenities);

-- Populate base_price_cents from decimal
UPDATE unit_types SET base_price_cents = (base_price * 100)::BIGINT WHERE base_price_cents = 0;

-- Amenities catalog (global)
CREATE TABLE amenities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
CREATE INDEX idx_amenities_name ON amenities(name);

-- Guest Services (renamed from hotel_services)
CREATE TABLE guest_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    icon VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
CREATE TRIGGER update_guest_services_modtime 
    BEFORE UPDATE ON guest_services 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE INDEX idx_guest_services_name ON guest_services(name);

-- Units (physical rooms)
CREATE TABLE units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL,
    unit_type_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_unit_type_id ON units(unit_type_id);
