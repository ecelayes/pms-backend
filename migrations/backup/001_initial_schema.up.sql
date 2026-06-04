CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

CREATE TABLE room_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    total_quantity INT NOT NULL CHECK (total_quantity >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE price_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    validity_range DATERANGE NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    priority INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_range CHECK (NOT isempty(validity_range))
);
CREATE INDEX idx_price_rules_range ON price_rules USING GIST (room_type_id, validity_range);

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    stay_range DATERANGE NOT NULL, -- [CheckIn, CheckOut)
    guest_email TEXT NOT NULL,
    total_price DECIMAL(10, 2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'confirmed',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT valid_stay CHECK (NOT isempty(stay_range))
);
CREATE INDEX idx_reservations_overlap ON reservations USING GIST (room_type_id, stay_range);
