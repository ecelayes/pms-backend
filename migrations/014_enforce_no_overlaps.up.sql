CREATE EXTENSION IF NOT EXISTS btree_gist;

TRUNCATE price_rules;

ALTER TABLE price_rules DROP COLUMN IF EXISTS priority;

ALTER TABLE price_rules
ADD CONSTRAINT no_overlapping_prices 
EXCLUDE USING GIST (
    room_type_id WITH =, 
    validity_range WITH &&
) WHERE (deleted_at IS NULL);
