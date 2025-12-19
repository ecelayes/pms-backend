ALTER TABLE hotels RENAME TO properties;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='properties' AND column_name='type') THEN
        ALTER TABLE properties ADD COLUMN type VARCHAR(50) NOT NULL DEFAULT 'HOTEL';
    END IF;
END $$;
