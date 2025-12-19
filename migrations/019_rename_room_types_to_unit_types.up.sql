-- Rename table
ALTER TABLE room_types RENAME TO unit_types;

-- Rename columns in unit_types
ALTER TABLE unit_types RENAME COLUMN hotel_id TO property_id;

-- Rename columns in reservations
ALTER TABLE reservations RENAME COLUMN room_type_id TO unit_type_id;

-- Rename columns in rate_plans
ALTER TABLE rate_plans RENAME COLUMN hotel_id TO property_id;
ALTER TABLE rate_plans RENAME COLUMN room_type_id TO unit_type_id;

-- Rename columns in price_rules
ALTER TABLE price_rules RENAME COLUMN room_type_id TO unit_type_id;
