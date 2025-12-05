CREATE TABLE guests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    phone TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX idx_guests_email ON guests(email);
CREATE TRIGGER update_guests_modtime BEFORE UPDATE ON guests FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

ALTER TABLE reservations ADD COLUMN guest_id UUID REFERENCES guests(id);

ALTER TABLE reservations DROP COLUMN guest_email;
