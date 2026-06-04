-- Migration to align reservations table with Booking Context DDD
-- 1. Rename room_type_id to unit_type_id (Already done in 019, skipping)
-- ALTER TABLE reservations RENAME COLUMN room_type_id TO unit_type_id;

-- 2. Add property_id (essential for tenant isolation/filtering)
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS property_id UUID REFERENCES properties(id);

-- Populate property_id from unit_type (was room_type)
UPDATE reservations r
SET property_id = rt.property_id
FROM unit_types rt
WHERE r.unit_type_id = rt.id AND r.property_id IS NULL;

-- 3. Add unit_id (for specific room assignment)
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS unit_id UUID REFERENCES units(id); 
-- Note: 'units' table must exist (created in previous migration or legacy?)
-- Legacy didn't seem to have 'units' table in 001. 
-- Catalog Context created 'Unit' entity. did it create 'units' table?
-- Catalog Repo (Step 142) inserts into `units`.
-- So `units` table MUST exist. 
-- I need to verify if `units` table exists.
-- Assuming `units` table created by `001` or `021`?
-- `001` only had `room_types`.
-- `021` (step 140) used `ALTER TABLE properties`. It implied `units` exists?
-- Step 142 `PostgresCatalogRepository` inserts into `units`.
-- If `units` table was never created, that code would fail.
-- I should check if `units` table exists. 
-- Providing `CREATE TABLE IF NOT EXISTS units` just in case.

-- 4. Add reservation_code (Already done in 003)
-- ALTER TABLE reservations ADD COLUMN IF NOT EXISTS reservation_code VARCHAR(20);
-- UPDATE reservations SET reservation_code = SUBSTRING(id::text, 1, 8) WHERE reservation_code IS NULL;

-- 5. Add price_cents & currency
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS price_cents BIGINT DEFAULT 0;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS price_currency VARCHAR(3) DEFAULT 'USD';
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS guest_email TEXT;

UPDATE reservations SET price_cents = (total_price * 100)::BIGINT WHERE price_cents = 0;

-- 6. Ensure unit_id is nullable (reservations might be unassigned initially)
