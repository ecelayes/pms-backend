ALTER TABLE room_types 
ADD COLUMN base_price DECIMAL(10, 2) NOT NULL DEFAULT 0;

ALTER TABLE room_types ADD CONSTRAINT check_base_price_positive CHECK (base_price >= 0);
