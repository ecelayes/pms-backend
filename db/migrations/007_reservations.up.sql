-- Reservations and Price Rules

-- Price Rules with exclusion constraint
CREATE TABLE price_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unit_type_id UUID NOT NULL REFERENCES unit_types(id),
    validity_range DATERANGE NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    price_cents BIGINT DEFAULT 0,
    price_currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT valid_range CHECK (NOT isempty(validity_range))
);
CREATE INDEX idx_price_rules_range ON price_rules USING GIST (unit_type_id, validity_range);

-- Populate price_cents from decimal
UPDATE price_rules SET price_cents = (price * 100)::BIGINT WHERE price_cents = 0;

-- Add exclusion constraint (only if no overlapping data)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'no_overlapping_prices'
    ) THEN
        ALTER TABLE price_rules 
        ADD CONSTRAINT no_overlapping_prices 
        EXCLUDE USING GIST (
            unit_type_id WITH =, 
            validity_range WITH &&
        ) WHERE (deleted_at IS NULL);
    END IF;
EXCEPTION WHEN undefined_object THEN
    -- Constraint already exists or other error, skip
    NULL;
END $$;

-- Reservations table
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    unit_type_id UUID NOT NULL REFERENCES unit_types(id),
    unit_id UUID REFERENCES units(id),
    guest_id UUID REFERENCES guests(id),
    rate_plan_id UUID REFERENCES rate_plans(id),
    stay_range DATERANGE NOT NULL,
    reservation_code TEXT UNIQUE,
    adults INT NOT NULL DEFAULT 1,
    children INT NOT NULL DEFAULT 0,
    total_price DECIMAL(10, 2) NOT NULL,
    price_cents BIGINT DEFAULT 0,
    price_currency VARCHAR(3) DEFAULT 'USD',
    status TEXT NOT NULL DEFAULT 'confirmed',
    guest_email TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT valid_stay CHECK (NOT isempty(stay_range))
);
CREATE TRIGGER update_reservations_modtime 
    BEFORE UPDATE ON reservations 
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE INDEX idx_reservations_overlap ON reservations USING GIST (unit_type_id, stay_range);
CREATE INDEX idx_reservations_guest_id ON reservations(guest_id);
CREATE INDEX idx_reservations_rate_plan ON reservations(rate_plan_id);
CREATE INDEX idx_reservations_deleted_at ON reservations(deleted_at) WHERE deleted_at IS NULL;

-- Populate price_cents from decimal
UPDATE reservations SET price_cents = (total_price * 100)::BIGINT WHERE price_cents = 0;
