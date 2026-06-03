-- Migration to support new Catalog Domain strict types
-- 1. Add base_price_cents to unit_types (refactoring float money)
ALTER TABLE unit_types ADD COLUMN IF NOT EXISTS base_price_cents BIGINT DEFAULT 0;
ALTER TABLE unit_types ADD COLUMN IF NOT EXISTS base_price_currency VARCHAR(3) DEFAULT 'USD';

-- Populate cents from old base_price (decimal)
UPDATE unit_types SET base_price_cents = (base_price * 100)::BIGINT WHERE base_price_cents = 0;

-- 2. Add type to properties (Already done in 018)
-- ALTER TABLE properties ADD COLUMN IF NOT EXISTS type VARCHAR(50) DEFAULT 'hotel';
