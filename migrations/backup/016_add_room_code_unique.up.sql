ALTER TABLE room_types ALTER COLUMN code SET NOT NULL;
ALTER TABLE room_types ADD CONSTRAINT unique_room_code_per_hotel UNIQUE (hotel_id, code);
