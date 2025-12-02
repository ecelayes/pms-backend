TRUNCATE reservations, price_rules, room_types CASCADE;

INSERT INTO room_types (id, name, total_quantity) 
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Deluxe Ocean View', 5),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'Standard Garden', 10);

INSERT INTO price_rules (room_type_id, validity_range, price, priority)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '[2024-01-01, 2025-12-31)', 200.00, 0),
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', '[2024-01-01, 2025-12-31)', 100.00, 0);

INSERT INTO price_rules (room_type_id, validity_range, price, priority)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '[2024-07-01, 2024-09-01)', 350.00, 10);

INSERT INTO price_rules (room_type_id, validity_range, price, priority)
VALUES 
    ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '[2024-05-10, 2024-05-13)', 150.00, 100);
