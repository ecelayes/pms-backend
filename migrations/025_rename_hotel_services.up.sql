-- Rename table hotel_services to guest_services
ALTER TABLE hotel_services RENAME TO guest_services;

-- Rename index if exists
ALTER INDEX IF EXISTS idx_hotel_services_name RENAME TO idx_guest_services_name;
