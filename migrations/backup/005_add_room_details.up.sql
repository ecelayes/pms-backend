ALTER TABLE room_types 
ADD COLUMN max_occupancy INT NOT NULL DEFAULT 2,
ADD COLUMN amenities TEXT[];

UPDATE room_types SET amenities = ARRAY['wifi', 'tv', 'ac', 'shower'] WHERE amenities IS NULL;
