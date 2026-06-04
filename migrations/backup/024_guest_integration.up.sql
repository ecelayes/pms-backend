-- Migration 024: Link Reservations to Users (Guests)
-- 1. Add guest_id column (Already done in 007, referencing guests)
-- ALTER TABLE reservations ADD COLUMN IF NOT EXISTS guest_id UUID REFERENCES users(id);

-- 2. Populate guest_id (Skip if column exists, old logic)
-- UPDATE reservations r
-- SET guest_id = u.id
-- FROM users u
-- WHERE r.guest_email = u.email AND r.guest_id IS NULL;
