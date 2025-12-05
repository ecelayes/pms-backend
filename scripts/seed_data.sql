TRUNCATE reservations, price_rules, room_types, hotels, users CASCADE;

INSERT INTO users (id, email, password, salt, role) 
VALUES (
    '10eebc99-9c0b-4ef8-bb6d-6bb9bd380a00',
    'admin@hotel.com', 
    '$2a$10$wS.xX/qY.D.I/..passwordhash..', 
    'salt_placeholder', 
    'admin'
);

INSERT INTO hotels (id, owner_id, name, code)
VALUES (
    '20eebc99-9c0b-4ef8-bb6d-6bb9bd380a99',
    '10eebc99-9c0b-4ef8-bb6d-6bb9bd380a00',
    'Grand Hotel Miami',
    'MIA'
);

INSERT INTO room_types (id, hotel_id, name, code, total_quantity) 
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '20eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'Presidential Suite', 'SUI', 2),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', '20eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'Deluxe Ocean View', 'DLX', 10);

INSERT INTO price_rules (room_type_id, validity_range, price, priority)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '[2024-01-01, 2025-12-31)', 500.00, 0),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', '[2024-01-01, 2025-12-31)', 200.00, 0);

INSERT INTO guests (id, email, first_name, last_name) 
VALUES ('g0eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'vip@guest.com', 'Lionel', 'Messi');
