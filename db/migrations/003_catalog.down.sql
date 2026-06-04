DROP TRIGGER IF EXISTS update_unit_types_modtime ON unit_types;
DROP TRIGGER IF EXISTS update_guest_services_modtime ON guest_services;
DROP TRIGGER IF EXISTS update_amenities_modtime ON amenities;

DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS guest_services;
DROP TABLE IF EXISTS amenities;
DROP TABLE IF EXISTS unit_types;
