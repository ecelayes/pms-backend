CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    code VARCHAR(5) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TRIGGER update_hotels_modtime BEFORE UPDATE ON hotels FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

ALTER TABLE room_types ADD COLUMN hotel_id UUID REFERENCES hotels(id);
ALTER TABLE room_types ADD COLUMN code VARCHAR(5) DEFAULT 'STD';
ALTER TABLE room_types ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
CREATE TRIGGER update_room_types_modtime BEFORE UPDATE ON room_types FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

ALTER TABLE reservations ADD COLUMN reservation_code TEXT UNIQUE;
ALTER TABLE reservations ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
CREATE TRIGGER update_reservations_modtime BEFORE UPDATE ON reservations FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

ALTER TABLE users ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
CREATE TRIGGER update_users_modtime BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

ALTER TABLE price_rules ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
CREATE TRIGGER update_price_rules_modtime BEFORE UPDATE ON price_rules FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
