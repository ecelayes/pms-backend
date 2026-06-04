DROP TRIGGER IF EXISTS update_reservations_modtime ON reservations;
DROP TRIGGER IF EXISTS no_overlapping_prices ON price_rules;

DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS price_rules;
