ALTER TABLE room_types 
ADD COLUMN max_adults INT NOT NULL DEFAULT 2,
ADD COLUMN max_children INT NOT NULL DEFAULT 0;

UPDATE room_types SET max_adults = max_occupancy;
