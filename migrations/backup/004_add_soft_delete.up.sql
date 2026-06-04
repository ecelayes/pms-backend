ALTER TABLE users ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE hotels ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE room_types ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE price_rules ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;
ALTER TABLE reservations ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_hotels_deleted_at ON hotels(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_room_types_deleted_at ON room_types(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_price_rules_deleted_at ON price_rules(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservations_deleted_at ON reservations(deleted_at) WHERE deleted_at IS NULL;
