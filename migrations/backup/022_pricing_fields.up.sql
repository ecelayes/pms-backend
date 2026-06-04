-- Migration to add price_cents to price_rules
ALTER TABLE price_rules ADD COLUMN IF NOT EXISTS price_cents BIGINT DEFAULT 0;
ALTER TABLE price_rules ADD COLUMN IF NOT EXISTS price_currency VARCHAR(3) DEFAULT 'USD';

UPDATE price_rules SET price_cents = (price * 100)::BIGINT WHERE price_cents = 0;
